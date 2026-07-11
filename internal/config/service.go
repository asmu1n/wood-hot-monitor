package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type ConfigCallback func(cfg AppConfig)

// 本地文件配置，读写锁保护
type Service struct {
	path            string
	mu              sync.RWMutex
	cfg             AppConfig
	updateListeners map[uint64]ConfigCallback
	cbId            uint64
}

func NewService(dataDir string) (*Service, error) {
	s := &Service{
		path:            filepath.Join(dataDir, "config.json"),
		cfg:             defaultConfig(),
		updateListeners: make(map[uint64]ConfigCallback),
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func (s *Service) Get() (*AppConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.cfg
	return &cfg, nil
}

func (s *Service) Update(cfg AppConfig) error {
	s.mu.Lock()
	defer func() {
		curCfg := s.cfg
		s.mu.Unlock()
		s.notify(&curCfg)
	}()
	s.cfg = cfg

	return s.save()
}

func (s *Service) GetSetting(key string) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.Settings[key], nil
}

func (s *Service) UpdateSettings(settings map[string]any) error {
	s.mu.Lock()
	defer func() {
		curCfg := s.cfg
		s.mu.Unlock()
		s.notify(&curCfg)
	}()
	if s.cfg.Settings == nil {
		s.cfg.Settings = make(map[string]any)
	}
	for k, v := range settings {
		s.cfg.Settings[k] = v
	}

	return s.save()
}

func (s *Service) SubscribeUpdates(cb func(AppConfig)) func() {
	s.mu.Lock()
	s.cbId++
	id := s.cbId
	s.updateListeners[id] = cb
	s.mu.Unlock()
	unSub := func() {
		s.mu.Lock()
		delete(s.updateListeners, id)
		s.mu.Unlock()
	}

	return unSub
}

func (s *Service) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.cfg)
}

func (s *Service) save() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Service) notify(val *AppConfig) {
	s.mu.RLock()

	listeners := make([]ConfigCallback, 0, len(s.updateListeners))

	for _, cb := range s.updateListeners {
		listeners = append(listeners, cb)

	}

	s.mu.RUnlock()

	for _, cb := range listeners {
		cb(*val)
	}
}

func defaultConfig() AppConfig {
	return AppConfig{
		CheckInterval: 30,
		Settings:      make(map[string]any),
	}
}
