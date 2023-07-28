package acmd

import (
	"runtime"
	"runtime/debug"
	"time"
)

// BuildInfo of the application. Helper for [debug.ReadBuildInfo].
type BuildInfo struct {
	GoVersion  string
	Version    string
	Revision   string
	LastCommit time.Time
	DirtyBuild bool
}

func (bi *BuildInfo) String() string {
	return bi.Version + "-" + bi.Revision
}

// Version returns the BuildInfo of the application.
func Version() BuildInfo {
	info := BuildInfo{
		GoVersion:  runtime.Version(),
		Revision:   "SNAPSHOT",
		DirtyBuild: true,
	}

	debugInfo, ok := debug.ReadBuildInfo()
	if !ok {
		info.Version = debugInfo.Main.Version
		return info
	}

	for _, kv := range debugInfo.Settings {
		switch kv.Key {
		case "vcs.revision":
			info.Revision = kv.Value
		case "vcs.time":
			info.LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			info.DirtyBuild = kv.Value == "true"
		}
	}
	return info
}
