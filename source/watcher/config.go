package watcher

import (
	"strings"
	"time"
)

type Entry struct {
	Timeout   *time.Duration            `yaml:"timeout"`
	Overrides map[string]*time.Duration `yaml:"overrides"`
	DeferSync *bool                     `yaml:"deferSync"`
}

type Config struct {
	Timeout      time.Duration    `yaml:"timeout"`
	DeferSync    bool             `yaml:"deferSync"`
	Applications map[string]Entry `yaml:"applications"`
}

func (c *Config) GetTimeout(app, cell string) time.Duration {
	for appOverride, appOverrideCfg := range c.Applications {
		if strings.HasPrefix(app, appOverride) {
			for cellOverride, cellOverrideTimeout := range appOverrideCfg.Overrides {
				if strings.HasPrefix(cell, cellOverride) {
					if cellOverrideTimeout != nil {
						return *cellOverrideTimeout
					}
					break
				}
			}
			if appOverrideCfg.Timeout != nil {
				return *appOverrideCfg.Timeout
			}
			break
		}
	}
	return c.Timeout
}

func (c *Config) GetDeferSync(app string) bool {
	for appOverride, appOverrideCfg := range c.Applications {
		if strings.HasPrefix(app, appOverride) {
			if appOverrideCfg.DeferSync != nil {
				return *appOverrideCfg.DeferSync
			}
			break
		}
	}
	return c.DeferSync
}
