package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"swarm"
)

// Artisan struct for Laravel artisan commands execution.
type Artisan struct {
	base string
	args []string
}

// NewArtisan creates Artisan struct with base arguments definition.
func NewArtisan() *Artisan {
	configPath := swarm.Config().ArtisanPath
	var args []string
	var baseExec string

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		for _, i := range strings.Split(configPath, " ") {
			args = append(args, i)
		}

		for _, j := range []string{"php", "artisan"} {
			args = append(args, j)
		}

		if len(args) > 1 {
			baseExec, args = args[0], args[1:]
		}
	} else {
		baseExec = "php"
		artisanPath := fmt.Sprintf("%s/%s", configPath, "artisan")
		args = []string{artisanPath}
	}

	return &Artisan{
		base: baseExec,
		args: args,
	}
}

// Exec runs artisan command returning its final output.
func (a *Artisan) Exec(args ...string) ([]byte, *exec.Cmd, error) {
	cmd := exec.Command(a.base, a.parseArgs(args)...)

	output, err := cmd.Output()
	return output, cmd, err
}

// ExecPipe runs artisan command constantly gathering all its output if the command is still running.
// Usable by listeners/queues.
func (a *Artisan) ExecPipe(handler func(output string, cms *exec.Cmd) error, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(a.base, a.parseArgs(args)...)

	stdout, err := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	buff := make([]byte, 1024)
	var n int
	for err == nil {
		n, err = stdout.Read(buff)
		if n > 0 {
			err := handler(string(buff[:n]), cmd)
			if err != nil {
				break
			}
		}
	}

	_ = cmd.Wait()

	return cmd, nil
}

// parseArgs adding custom args to the defined ones.
func (a *Artisan) parseArgs(args []string) []string {
	execArgs := a.args
	for _, i := range args {
		execArgs = append(execArgs, i)
	}

	return execArgs
}
