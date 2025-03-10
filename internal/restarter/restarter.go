package restarter

import (
	"fmt"
	plugin "github.com/k4ties/df-plugin/df-plugin"
	"os"
	"os/exec"
	"sync"
	"time"
)

var doOnRestart = struct {
	f   []func()
	fMu sync.Mutex
}{}

func DoOnRestart(f func()) {
	doOnRestart.fMu.Lock()
	doOnRestart.f = append(doOnRestart.f, f)
	doOnRestart.fMu.Unlock()
}

var cmd = exec.Command(mustExecutable())

func init() {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func mustExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return exe
}

func Restart() error {
	for p := range plugin.M().Srv().Players(nil) {
		_ = p.Transfer(p.Data().Session.ClientData().ServerAddress)
	}

	<-time.After(time.Second)
	for _, f := range doOnRestart.f {
		f()
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to restart: %w", err)
	}

	os.Exit(0)
	return nil
}
