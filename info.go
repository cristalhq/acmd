package acmd

import (
	"runtime/debug"
	"time"
)

// BuildInfo of the application. Helper for [debug.ReadBuildInfo].
type BuildInfo struct {
	Revision     string
	LastCommit   time.Time
	IsDirtyBuild bool
}

// GetBuildInfo returns build information from [debug.ReadBuildInfo].
func GetBuildInfo() BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return BuildInfo{
			Revision:     "SNAPSHOT",
			IsDirtyBuild: true,
		}
	}

	var bi BuildInfo

	for _, kv := range info.Settings {
		if kv.Value == "" {
			continue
		}

		switch kv.Key {
		case "vcs.revision":
			bi.Revision = kv.Value
		case "vcs.time":
			bi.LastCommit, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			bi.IsDirtyBuild = kv.Value == "true"
		}
	}
	return bi
}
