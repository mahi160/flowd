package fw

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

type Platform struct {
	Hostname string
	OS       string
	Arch     string
	Machine  string
}

var (
	platform     *Platform
	platformOnce sync.Once
)

// initPlatform initialises the singleton Platform. Must be called once at
// daemon start before any goroutine calls GetPlatform(). machineName is the
// user-configured name; falls back to the short hostname when empty.
func initPlatform(machineName string) {
	platformOnce.Do(func() {
		host, _ := os.Hostname()
		short := strings.SplitN(host, ".", 2)[0]
		name := machineName
		if name == "" {
			name = short
		}
		platform = &Platform{
			Hostname: short,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			Machine:  name,
		}
	})
}

// ScreenClosed reports whether the laptop lid is physically closed.
// Only implemented on macOS (via IOKit's AppleClamshellState); returns false
// on all other platforms so tracking continues normally.
func ScreenClosed() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	out, err := exec.Command("ioreg", "-r", "-k", "AppleClamshellState", "-d", "4").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), `"AppleClamshellState" = Yes`)
}

func GetPlatform() *Platform {
	if platform == nil {
		initPlatform("")
	}
	return platform
}
