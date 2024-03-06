package runner

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"sync"
)

var _mx sync.Mutex

func RunCommand(pathToCommand string, args ...string) ([]byte, int, error) {
	_mx.Lock()
	defer _mx.Unlock()

	argv0, err := exec.LookPath(pathToCommand)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to find command: %w", err)
	}

	cmd := exec.Command(argv0, args...)
	statusCode := 0
	var out []byte
	if out, err = cmd.CombinedOutput(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			statusCode = exitErr.ExitCode()
		}
	}
	return out, statusCode, err
}
