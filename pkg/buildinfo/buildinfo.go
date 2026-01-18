package buildinfo

import (
	"os"
	"runtime"
	"time"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
	StartedAt = time.Now()
	hostname  string
)

func init() {
	hostname, _ = os.Hostname()
}

func Hostname() string      { return hostname }
func GoVersion() string     { return runtime.Version() }
func Uptime() time.Duration { return time.Since(StartedAt) }
