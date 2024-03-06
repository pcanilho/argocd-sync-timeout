package main

import (
	"argocd-sync-timeout/argo"
	"argocd-sync-timeout/probes"
	"argocd-sync-timeout/watcher"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	_appPeriod time.Duration
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	p, err := time.ParseDuration(os.Getenv("AST_PERIOD"))
	if err != nil {
		logger.Error("Error parsing period from variable [AST_PERIOD]...", "error", err)
		os.Exit(1)
	}
	_appPeriod = p

	fw, cfg, err := watcher.NewFileWatcher(os.Getenv("AST_CONFIG"))
	if err != nil {
		logger.Error("Error creating file watcher from variable [AST_CONFIG]...", "error", err)
		os.Exit(1)
	}

	logger.Info("Loaded configuration...")

	changed := make(chan struct{}, 1)

	logger.Debug("Starting watcher...")
	go fw.Watch(changed)
	logger.Debug("Starting probes...")
	go probes.Run(logger)
	logger.Debug("Logging into ArgoCD...")
	if err = argo.Login(); err != nil {
		logger.Error("Error logging into ArgoCD...", "error", err)
		os.Exit(1)
	}

	for {
		select {
		case <-changed:
			logger.Info("Configuration changed...", "config", fmt.Sprintf("%+v", cfg))
		default:
			// Process cfg
			logger.Info(strings.Repeat("-", 80))
			logger.Info("Processing configuration...")
			apps, err := argo.ListApps()
			if err != nil {
				logger.Error("Error listing application...", "error", err)
				os.Exit(1)
			}
			if apps == nil || len(apps) == 0 {
				logger.Debug("No applications found...")
				continue
			}
			_wg := sync.WaitGroup{}
			for _, app := range apps {
				timeout := cfg.GetTimeout(app.Metadata.Name, app.Cell())
				deferSync := cfg.GetDeferSync(app.Metadata.Name)

				go func(appName string, appTimeout time.Duration) {
					_wg.Add(1)
					defer _wg.Done()
					logger.Debug("Processing application...", "application", appName, "timeout", appTimeout.String())
					opErr := argo.EnforceSyncTimeout(logger, appName, appTimeout, deferSync)
					if opErr != nil {
						logger.Error("Error processing operation...", "application", appName, "error", opErr)
					} else {
						logger.Debug("Finished processing application...", "application", appName)
					}
				}(app.Metadata.Name, timeout)
			}
			logger.Debug(strings.Repeat("-", 40))
			logger.Debug("Waiting for all applications to be processed...")
			_wg.Wait()
			if _appPeriod > 0 {
				logger.Info("Sleeping...", "period", _appPeriod.String())
				logger.Info(strings.Repeat("-", 80))
				<-time.After(_appPeriod)
			}
		}
	}
}
