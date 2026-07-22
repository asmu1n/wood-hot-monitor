package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/module/checker"
	"wood-hot-monitor/internal/module/hotspot"
	"wood-hot-monitor/internal/module/keyword"

	hsrepo "wood-hot-monitor/internal/module/hotspot/repository"
	kwrepo "wood-hot-monitor/internal/module/keyword/repository"

	"wood-hot-monitor/internal/infra/database"
	"wood-hot-monitor/internal/infra/llm"
	"wood-hot-monitor/internal/infra/notify"
	"wood-hot-monitor/internal/infra/scraper"

	"github.com/go-co-op/gocron/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dataDir := defaultDataDir()

	db, err := database.New(dataDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	cfgService, err := config.NewService(dataDir)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	hotspotRepo := hsrepo.NewHotspotRepository(db.Client)
	keywordRepo := kwrepo.NewKeywordRepository(db.Client)

	hotspotService := hotspot.NewService(hotspotRepo)
	keywordService := keyword.NewService(keywordRepo)

	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Services: []application.Service{
			application.NewService(keywordService),
			application.NewService(hotspotService),
			application.NewService(cfgService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	scraperService := scraper.NewService()
	notifier := notify.NewWailsNotifier(app)
	llmService := llm.NewService(cfgService, db.Client)
	checkerService := checker.NewService(keywordService, cfgService, llmService, hotspotService, scraperService, notifier)
	app.RegisterService(application.NewService(checkerService))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Wood Hot Monitor",
		Width:  1280,
		Height: 860,
		URL:    "/",
	})

	scheduler, err := startScheduler(cfgService, checkerService)
	if err != nil {
		log.Printf("warning: scheduler start failed: %v", err)
	} else {
		defer scheduler.Shutdown()
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func startScheduler(cfgService *config.Service, checkerService *checker.Service) (gocron.Scheduler, error) {
	cfg, err := cfgService.Get()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	interval := cfg.CheckInterval
	if interval <= 0 {
		interval = 30
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	job, err := s.NewJob(
		gocron.DurationJob(time.Duration(interval)*time.Minute),
		gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := checkerService.Run(ctx); err != nil {
				log.Printf("checker run error: %v", err)
			}
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("register job: %w", err)
	}

	cfgService.SubscribeUpdates(func(ac config.AppConfig) {
		updatedJob, err := s.Update(job.ID(), gocron.DurationJob(time.Duration(ac.CheckInterval)*time.Minute), gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := checkerService.Run(ctx); err != nil {
				log.Printf("checker run error: %v", err)
			}
		}))
		if err != nil {
			log.Printf("update job error: %v", err)
		}
		log.Printf("job updated: %v", updatedJob)
	})

	s.Start()

	log.Printf("scheduler started: check every %d minutes", interval)
	return s, nil
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".wood-hot-monitor")
}
