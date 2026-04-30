package fw

import (
	"os"
	"runtime"
	"strings"
)

type Platform struct {
	Hostname string
	OS       string
	Arch     string
	Machine  string // config override or hostname
}

var platform *Platform

func initPlatform(machineName string) {
	host, _ := os.Hostname()
	// strip .local / domain suffix for cleanliness
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
}

func GetPlatform() *Platform {
	if platform == nil {
		initPlatform("")
	}
	return platform
}
