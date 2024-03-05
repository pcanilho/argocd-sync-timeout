package runner

import (
	"github.com/pkg/errors"
	"os/exec"
)

func RunCommand(pathToCommand string, args ...string) ([]byte, int, error) {
	argv0, err := exec.LookPath(pathToCommand)
	if err != nil {
		return nil, -1, err
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
	return out, statusCode, nil
}
