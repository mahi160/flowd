package fw

import (
	"os"
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

func GetPlatform() *Platform {
	if platform == nil {
		initPlatform("")
	}
	return platform
}
