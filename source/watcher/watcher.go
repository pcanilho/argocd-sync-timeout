package watcher

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v3"
	"os"
	"sync"
)

type fileWatcher struct {
	watcher      *fsnotify.Watcher
	cfgSingleton Config
	mutex        sync.RWMutex
}

func NewFileWatcher(filePath string) (*fileWatcher, *Config, error) {
	var inst fileWatcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating new watcher: %v", err)
	}

	if err = watcher.Add(filePath); err != nil {
		return nil, nil, fmt.Errorf("error watching file: %v", err)
	}

	if err = inst.loadConfig(filePath); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading configuration: %v", err)
		os.Exit(1)
	}
	inst.mutex = sync.RWMutex{}
	inst.watcher = watcher
	return &inst, &inst.cfgSingleton, nil
}

func (w *fileWatcher) Watch(changed chan struct{}) {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := w.loadConfig(event.Name); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error reloading configuration: %v", err)
					os.Exit(1)
				}
				changed <- struct{}{}
			}
		case err := <-w.watcher.Errors:
			_, _ = fmt.Fprintf(os.Stderr, "Error watching file: %v", err)
			os.Exit(1)
		}
	}
}

func (w *fileWatcher) Close() error {
	err := w.watcher.Close()
	if err != nil {
		return fmt.Errorf("error closing watcher: %v", err)
	}
	return nil
}

func (w *fileWatcher) loadConfig(filePath string) error {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal YAML data into Config struct
	w.mutex.Lock()
	err = yaml.Unmarshal(data, &w.cfgSingleton)
	if err != nil {
		return fmt.Errorf("error unmarshalling YAML: %v", err)
	}
	w.mutex.Unlock()

	return nil
}
