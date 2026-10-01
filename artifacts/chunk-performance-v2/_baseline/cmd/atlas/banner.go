package main

import "log/slog"

// buildTime is set by build.ps1 with -ldflags. An ordinary go build leaves
// the value explicit instead of confusing the process start time with it.
var buildTime = "unknown (build without timestamp injection)"

var atlasArt = [...]string{
	`   _  _____ _        _    ____ `,
	`  / \|_   _| |      / \  / ___|`,
	` / _ \ | | | |     / _ \ \___ \`,
	`/ ___ \| | | |___ / ___ \ ___) |`,
	`/_/   \_\_| |_____/_/   \_\____/`,
}

func logStartupBanner() {
	for _, line := range atlasArt {
		slog.Info(line)
	}
	slog.Info("Atlas Minecraft Java 1.21.1 server")
	slog.Info("Build time: " + buildTime)
}
