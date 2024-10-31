package main

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/sagoresarker/nomad-filewatcher-driver/pkg/driver"
)

func main() {
	// Create logger
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Info,
		JSONFormat: true,
		Name:       "filewatcher-driver",
	})

	// Create and serve the plugin
	plugin := driver.NewFileWatcherDriver(logger)
	if err := drivers.Serve(plugin, logger); err != nil {
		logger.Error("failed to serve driver plugin", "error", err)
	}
}
