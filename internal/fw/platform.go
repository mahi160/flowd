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
	platformName string
)

func initPlatform(machineName string) {
	platformName = machineName
	platformOnce.Do(func() {
		host, _ := os.Hostname()
		short := strings.SplitN(host, ".", 2)[0]
		name := platformName
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
