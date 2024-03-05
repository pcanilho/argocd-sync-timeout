package main

import (
	"argocd-sync-timeout/argo"
	"argocd-sync-timeout/probes"
	"argocd-sync-timeout/watcher"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"
)

var (
	_appPeriod time.Duration
	_wg        sync.WaitGroup
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	p, err := time.ParseDuration(os.Getenv("AST_PERIOD"))
	if err != nil {
		logger.Error("Error parsing period...", "error", err)
		os.Exit(1)
	}
	_appPeriod = p

	fw, cfg, err := watcher.NewFileWatcher(os.Getenv("AST_CONFIG"))
	if err != nil {
		logger.Error("Error creating file watcher...", "error", err)
		os.Exit(1)
	}

	logger.Info("Loaded configuration...", "config", cfg)

	errorChan := make(chan error)
	changed := make(chan struct{}, 1)
	go fw.Watch(changed)
	go probes.Run(errorChan)

	logger.Info("Starting watcher...")
	for {
		select {
		case e := <-errorChan:
			logger.Error("Error processing configuration...", "error", e)
			os.Exit(1)
		case <-changed:
			logger.Info("Configuration changed...", "config", fmt.Sprintf("%+v", cfg))
		default:
			// Process cfg
			logger.Info("Processing configuration...")
			apps, err := argo.ListApps()
			if err != nil {
				logger.Error("Error listing apps...", "error", err)
				errorChan <- err
			}
			_wg.Add(len(apps))
			if apps == nil {
				logger.Debug("No applications found...")
				continue
			}
			for _, app := range apps {
				timeout := cfg.GetTimeout(app.Metadata.Name, app.Cell())
				deferSync := cfg.GetDeferSync(app.Metadata.Name)

				go func(appName string, appTimeout time.Duration) {
					logger.Debug("Processing application...", "app", appName, "timeout", appTimeout)
					opErr := argo.EnforceSyncTimeout(logger, appName, appTimeout, deferSync)
					if opErr != nil {
						logger.Error("Error processing operation...", "app", appName, "error", opErr)
						errorChan <- opErr
					}
					logger.Debug("Finished processing application...", "app", appName)
					_wg.Done()
				}(app.Metadata.Name, timeout)
			}
			_wg.Wait()
			if _appPeriod > 0 {
				<-time.After(_appPeriod)
			}
		}
	}
}
