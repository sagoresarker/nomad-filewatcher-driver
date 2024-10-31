package driver

import (
	"sync"
	"time"

	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/sagoresarker/nomad-filewatcher-driver/pkg/watcher"
)

type TaskHandle struct {
	mutex       sync.RWMutex
	taskConfig  *TaskConfig
	watcher     *watcher.FileWatcher
	exitResult  *drivers.ExitResult
	startedAt   time.Time
	completedAt time.Time
}

func (h *TaskHandle) TaskStatus() *drivers.TaskStatus {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.Paths[0],
		Name:        "filewatcher",
		State:       drivers.TaskStateRunning,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
	}
}
