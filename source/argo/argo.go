package argo

import (
	"argocd-sync-timeout/retrier"
	"argocd-sync-timeout/runner"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log/slog"
	"time"
)

const (
	_cli = "argocd"
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

func EnforceSyncTimeout(logger *slog.Logger, name string, timeout time.Duration, deferSync bool, retries int) (err error) {
	// Get app operation phase
	logger.Debug("Waiting for application operation to complete...", "application", name, "timeout", timeout.String())
	phase, code, err := retrier.RunWithRetryGetOperationStatus(logger, retries, getAppOperationStatus, name, timeout)
	if err != nil {
		logger.Error(errors.Wrapf(err, "[argo] failed to get application operation status. code: %d. phase: %s. error: %v", code, phase, err).Error())
		return errors.Wrapf(err, "[argo] failed to get application operation status. code: %d. phase: %s", code, phase)
	}
	logger.Debug("Processing application operation status...", "application", name, "phase", phase, "code", code, "error", err)
	if (phase == PhaseSucceeded || phase == PhaseRunning) && code == 0 {
		logger.Debug("Application operation completed successfully. Skipping...", "application", name, "code", code)
		return nil
	}
	// If the App sync is in error state, terminate the operation
	logger.Debug("Terminating application operation...", "application", name)
	_, code, err = retrier.RunWithRetryOperation(logger, retries, terminateAppOperation, name)
	if code != 0 && code != 20 {
		logger.Error(errors.Wrapf(err, "[argo] failed to terminate application operation. code: %d", code).Error(), "application", name)
		return errors.Wrapf(err, "[argo] failed to terminate application operation. code: %d", code)
	}

	if deferSync {
		logger.Debug("Launching application sync...", "application", name)
		err = retrier.RunWithRetryError(logger, retries, syncAppAsync, name)
		if err != nil {
			logger.Error(errors.Wrapf(err, "[argo] failed to launch application sync").Error(), "application", name)
		} else {
			logger.Debug("Application sync launched...", "application", name)
		}
	}

	return err
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
		return nil, errors.Wrap(err, "[argo] failed to unmarshal the ArgoCD application list")
	}
	return as, nil
}

func syncAppAsync(name string) error {
	out, code, err := runner.RunCommand(_cli, "app", "sync", name, "--prune", "--apply-out-of-sync-only", "--async", "--assumeYes", "--server-side", "--core")
	return errors.Wrapf(err, "[argo] failed to sync application. application: %v. code: %v. error: %v", name, code, string(out))
}

func getAppOperationStatus(name string, timeout time.Duration) (AppOperationPhase, int, error) {
	out, code, err := runner.RunCommand(_cli, "app", "wait", name, "--operation", "--timeout", fmt.Sprint(timeout.Seconds()), "--core")
	// Out of sync
	if code == 20 {
		return PhaseUnknown, code, nil
	}
	if err != nil {
		return PhaseFailed, code, errors.Wrapf(err, "[argo] failed to get application operation status. application: %v", name)
	}
	status, err := getAppOperationSyncStatusFromOutput(out)
	if status != SyncedStatus {
		return PhaseFailed, 0, nil
	}

	phase, err := getAppOperationPhaseFromOutput(out)
	return phase, code, err
}

func terminateAppOperation(name string) ([]byte, int, error) {
	out, code, err := runner.RunCommand(_cli, "app", "terminate-op", name, "--core")
	// No ongoing operation
	if code == 20 {
		return out, code, nil
	}
	if err != nil {
		return out, code, errors.Wrapf(err, "[argo] failed to terminate application operation. application: %v. output: %v", name, string(out))
	}
	return out, code, nil
}
