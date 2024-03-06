package argo

import (
	"argocd-sync-timeout/runner"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

const (
	_cli = "/bin/argocd"
)

type app struct {
	Metadata struct {
		Name   string `json:"name"`
		Labels struct {
			Account               string `json:"account"`
			AppKubernetesIoPartOf string `json:"app.kubernetes.io/part-of"`
			Cluster               string `json:"cluster"`
			Env                   string `json:"env"`
			Region                string `json:"region"`
		} `json:"labels"`
	} `json:"metadata"`
	Spec struct {
		Destination struct {
			Name string `json:"name"`
		}
	}
}

func (a *app) Cell() string {
	partOf := a.Metadata.Labels.AppKubernetesIoPartOf
	if partOf == "application-set" && a.Metadata.Labels.Env != "" && a.Metadata.Labels.Account != "" && a.Metadata.Labels.Region != "" && a.Metadata.Labels.Cluster != "" {
		return fmt.Sprintf("%s/%s/%s/%s", a.Metadata.Labels.Env, a.Metadata.Labels.Account, a.Metadata.Labels.Region, a.Metadata.Labels.Cluster)
	}
	return a.Spec.Destination.Name
}

type Apps = []app

func Login() error {
	_, _, err := runner.RunCommand(_cli, "login", "--core")
	return errors.Wrap(err, "[argo] failed to login")
}

func EnforceSyncTimeout(logger *slog.Logger, name string, timeout time.Duration, deferSync bool) (err error) {
	// Get app operation status
	logger.Debug("Waiting for application operation to complete...", "application", name, "timeout", timeout.String())
	out, code, err := getAppOperationStatus(name, timeout)
	if code == -1 {
		logger.Error(errors.Wrapf(err, "[argo] failed to get application operation status. code: %d. output: %s", code, string(out)).Error())
	}
	logger.Debug("Application operation status...", "application", name, "code", code, "error", err)
	if code == 0 {
		logger.Debug("Application operation completed successfully. Skipping...", "application", name, "code", code)
		return nil
	}
	// If the App sync is in error state, terminate the operation
	logger.Debug("Terminating application operation...", "application", name)
	_, code, err = terminateAppOperation(name)
	if code != 0 || err != nil {
		return errors.Wrap(err, "[argo] failed to terminate app operation")
	}
	if deferSync {
		logger.Debug("Launching application sync...", "application", name)
		err = syncAppAsync(name)
		if err != nil {
			logger.Error(errors.Wrapf(err, "[argo] failed to launch application sync").Error())
		} else {
			logger.Debug("Application sync launched...", "application", name)
		}
	}

	return nil
}

func ListApps() (Apps, error) {
	out, code, err := runner.RunCommand(_cli, "app", "list", "-o", "json", "--core")
	if err != nil {
		return nil, errors.Wrap(err, "[argo] failed to list Apps")
	}
	if code != 0 {
		return nil, errors.Errorf("[argo] failed to list Apps: %s", string(out))
	}
	var as Apps
	if err = json.Unmarshal(out, &as); err != nil {
		return nil, errors.Wrap(err, "[argo] failed to unmarshal Apps")
	}
	return as, nil
}

func syncAppAsync(name string) error {
	_, _, err := runner.RunCommand(_cli, "app", "sync", name, "--prune", "--apply-out-of-sync-only", "--async", "--core")
	return err
}

func getAppOperationStatus(name string, timeout time.Duration) ([]byte, int, error) {
	out, code, err := runner.RunCommand(_cli, "app", "wait", name, "--operation", "--timeout", fmt.Sprint(timeout.Seconds()), "--core")
	if err != nil {
		return nil, code, errors.Wrap(err, "[argo] failed to get app operation status")
	}
	return out, code, nil
}

func terminateAppOperation(name string) ([]byte, int, error) {
	_, code, err := runner.RunCommand(_cli, "app", "terminate-op", name, "--core")
	if err != nil {
		return nil, code, errors.Wrap(err, "[argo] failed to terminate app operation")
	}
	return nil, code, nil
}
