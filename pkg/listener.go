package pkg

import (
	"errors"
	"fmt"
	"greed"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Listener struct
type Listener struct {
	Listeners               map[string]*StreamListener
	newListenerHandlers     []func(name string)
	listenerChangedHandlers []func(listener StreamListener)
	artisan                 *Artisan
}

// NewListener creates listener with artisan command.
func NewListener() (*Listener, error) {
	artisan := NewArtisan()

	_, _, err := artisan.Exec("list", "streamer")
	if err != nil {
		return nil, errors.New("artisan not detected")
	}

	listener := &Listener{
		artisan: artisan,
	}

	return listener, nil
}

func (l *Listener) Listen(stream Stream) {
	lis := l.AddStreamListener(stream.Name)
	if lis.stopped {
		return
	}

	messages := stream.GetMessagesList()
	lastID := "0-0"
	if len(messages) > 0 {
		lastID = messages[len(messages)-1]
	}

	id := fmt.Sprintf("--last_id=%s", lastID)
	args := []string{"streamer:listen", stream.Name, id}

	for {
		cmd, err := l.artisan.ExecPipe(func(output string, cmd *exec.Cmd) error {
			if lis.stopped {
				return errors.New("stopped")
			}

			lis.Output = append(lis.Output, fmt.Sprintf("%s: %s", time.Now().Format("01-02-2006 15:04:05"), output))
			if lis.HasNoListeners(output) {
				lis.stopped = true
				l.emitListenerChanged(*lis)
				return errors.New("stopped")
			}

			if lis.IsFailing(output) {
				lis.warning = true
				l.emitListenerChanged(*lis)
			}

			return nil
		}, args...)

		if lis.stopped {
			return
		}

		if err != nil {
			return
		}

		code := cmd.ProcessState.ExitCode()
		if code == 1 {
			lis.error = true
			l.emitListenerChanged(*lis)
			args = []string{"streamer:listen", stream.Name}
			continue
		}
	}
}

func (l *Listener) AddStreamListener(name string) *StreamListener {
	if l.Listeners == nil {
		l.Listeners = make(map[string]*StreamListener)
	}

	lis, ok := l.Listeners[name]
	if ok {
		return lis
	}

	lis = &StreamListener{
		Name:   name,
		Output: nil,
	}

	l.Listeners[name] = lis
	l.emitNewListener(name)

	return lis
}

func (l *Listener) OnNewListener(handle func(a string)) {
	l.newListenerHandlers = append(l.newListenerHandlers, handle)
}

func (l *Listener) emitNewListener(name string) {
	for _, h := range l.newListenerHandlers {
		h(name)
	}
}

func (l *Listener) OnListenerChange(handle func(listener StreamListener)) {
	l.listenerChangedHandlers = append(l.listenerChangedHandlers, handle)
}

func (l *Listener) emitListenerChanged(listener StreamListener) {
	for _, h := range l.listenerChangedHandlers {
		h(listener)
	}
}

type StreamListener struct {
	Name    string
	Output  []string
	stopped bool
	warning bool
	error   bool
}

func (s StreamListener) ParseOutput() string {
	var content string
	for _, i := range s.Output {
		content += fmt.Sprintf("%s", i)
	}

	return content
}

func (s StreamListener) HasNoListeners(output string) bool {
	return output == fmt.Sprintf("There are no local listeners associated with %s event in configuration.\n", s.Name)
}

func (s StreamListener) IsFailing(output string) bool {
	return strings.Contains(output, "Listener error. Failed processing message")
}

func (s StreamListener) GetStatus() string {
	if s.error {
		return "[red]WARNING[red]"
	}

	if s.warning {
		return "[yellow]WARNING[yellow]"
	}

	if s.stopped {
		return "[grey]STOPPED[grey]"
	}

	return "[green]OK[green]"
}

// Artisan struct for Laravel artisan commands execution
type Artisan struct {
	Main   string
	args   []string
	config string
}

// NewArtisan creates new artisan command
func NewArtisan() *Artisan {
	configPath := greed.Config().ArtisanPath
	var args []string
	var mainExec string

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		for _, i := range strings.Split(configPath, " ") {
			args = append(args, i)
		}

		for _, j := range []string{"php", "artisan"} {
			args = append(args, j)
		}

		mainExec, args = args[0], args[1:]
	} else {
		mainExec = "php"
		artisanPath := fmt.Sprintf("%s/%s", configPath, "artisan")
		args = []string{artisanPath}
	}

	return &Artisan{
		Main:   mainExec,
		args:   args,
		config: configPath,
	}
}

// Exec runs artisan command
func (a *Artisan) Exec(args ...string) ([]byte, *exec.Cmd, error) {
	cmd := exec.Command(a.Main, a.parseArgs(args)...)

	output, err := cmd.Output()
	return output, cmd, err
}

// ExecPipe runs artisan command constantly gathering all its output if the command is still running.
// Usable by listeners/queues.
func (a *Artisan) ExecPipe(handle func(output string, cms *exec.Cmd) error, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(a.Main, a.parseArgs(args)...)

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
			err := handle(string(buff[:n]), cmd)
			if err != nil {
				break
			}
		}
	}

	_ = cmd.Wait()

	return cmd, nil
}

func (a *Artisan) parseArgs(args []string) []string {
	execArgs := a.args
	for _, i := range args {
		execArgs = append(execArgs, i)
	}

	return execArgs
}

func copyAndCapture(w io.Writer, r io.Reader) ([]byte, error) {
	var out []byte
	buf := make([]byte, 1024, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			out = append(out, d...)
			_, err := w.Write(d)
			if err != nil {
				return out, err
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return out, err
		}
	}
}
