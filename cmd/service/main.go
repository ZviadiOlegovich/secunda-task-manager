package main

import (
	"log/slog"
	"os"

	"github.com/zoshc/secunda-task-manager/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}
