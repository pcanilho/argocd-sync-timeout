package argo

import (
	"github.com/pkg/errors"
	"regexp"
)

var phaseRegex = regexp.MustCompile(`(?m)^Phase:\s+(.*)$`)
var statusRegex = regexp.MustCompile(`(?m)^Sync Status:\s+(.*)$`)

type AppOperationPhase = string
type AppOperationStatus = string

const (
	PhaseUnknown     AppOperationPhase = "Unknown"
	PhaseRunning                       = "Running"
	PhaseError                         = "Error"
	PhaseFailed                        = "Failed"
	PhaseSucceeded                     = "Succeeded"
	PhaseTerminating                   = "Terminating"
)

const (
	UnknownStatus   AppOperationStatus = "Unknown"
	SyncedStatus                       = "SyncedStatus"
	OutOfSyncStatus                    = "OutOfSyncStatus"
)

func getAppOperationPhaseFromOutput(out []byte) (AppOperationPhase, error) {
	phase := phaseRegex.FindAllSubmatch(out, 1)
	if len(phase) == 0 {
		return PhaseUnknown, errors.Errorf("[argo] failed to get application operation phase: %s", string(out))
	}
	return AppOperationPhase(phase[0][1]), nil
}

func getAppOperationSyncStatusFromOutput(out []byte) (AppOperationStatus, error) {
	status := statusRegex.FindAllSubmatch(out, 1)
	if len(status) == 0 {
		return UnknownStatus, errors.Errorf("[argo] failed to get application operation status: %s", string(out))
	}
	return AppOperationStatus(status[0][1]), nil
}
