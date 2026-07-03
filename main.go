package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"wood-hot-monitor/internal/biz/hotspot"
	"wood-hot-monitor/internal/biz/keyword"
	"wood-hot-monitor/internal/core/config"
	"wood-hot-monitor/internal/core/event"
	"wood-hot-monitor/internal/infra/database"
	"wood-hot-monitor/internal/infra/llm"
	"wood-hot-monitor/internal/worker/checker"

	"github.com/go-co-op/gocron/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. 初始化数据库、本地配置文件
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
	//  ----

	// 2. 初始化 Wails 应用，注入相应服务并运行应用
	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Services: []application.Service{
			application.NewService(keyword.NewService(db)),
			application.NewService(hotspot.NewService(db)),
			application.NewService(cfgService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	notifier := event.NewWailsNotifier(app)

	// 创建LLM服务
	llmService := llm.NewService(cfgService, db)
	checkerService := checker.NewService(db, cfgService, llmService, notifier)
	app.RegisterService(application.NewService(checkerService))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Wood Hot Monitor",
		Width:  1280,
		Height: 860,
		URL:    "/",
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
	// ----

	// 3.启动定时后台服务
	scheduler, err := startScheduler(cfgService, checkerService)
	if err != nil {
		log.Printf("warning: scheduler start failed: %v", err)
	} else {
		defer scheduler.Shutdown()
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

	_, err = s.NewJob(
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
