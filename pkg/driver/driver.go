package driver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/sagoresarker/nomad-filewatcher-driver/pkg/watcher"
)

const (
	pluginName    = "filewatcher-driver"
	pluginVersion = "v0.1.0"
)

type Driver struct {
	// This enables plugin clients to verify that an allocation belongs to the
	// expected task group.
	eventer        *eventer.Eventer
	config         *Config
	tasks          map[string]*TaskHandle
	ctx            context.Context
	signalShutdown context.CancelFunc
	logger         hclog.Logger
	lock           sync.RWMutex
}

func NewFileWatcherDriver(logger hclog.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)

	return &Driver{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		tasks:          make(map[string]*TaskHandle),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

func (d *Driver) PluginInfo() (*base.PluginInfoResponse, error) {
	return &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{"0.1.0"},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}, nil
}

func (d *Driver) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

func (d *Driver) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return fmt.Errorf("failed to decode driver config: %v", err)
		}
	}

	d.config = &config
	return nil
}

func (d *Driver) TaskConfigSchema() (*hclspec.Spec, error) {
	return taskConfigSpec, nil
}

func (d *Driver) Capabilities() (*drivers.Capabilities, error) {
	return &drivers.Capabilities{
		SendSignals: false,
		Exec:        false,
	}, nil
}

func (d *Driver) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	var taskConfig TaskConfig
	if err := cfg.DecodeDriverConfig(&taskConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to decode driver config: %v", err)
	}

	// Validate configuration
	if err := d.validateConfig(&taskConfig); err != nil {
		return nil, nil, fmt.Errorf("invalid config: %v", err)
	}

	// Create file watcher instance
	fw, err := watcher.NewFileWatcher(
		d.logger.Named(cfg.Name),
		taskConfig.Paths,
		taskConfig.Events,
		taskConfig.ExecCommand,
		taskConfig.ExecArgs,
		taskConfig.Environment,
		taskConfig.IgnorePatterns,
		taskConfig.RecursiveWatch,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create file watcher: %v", err)
	}

	// Start the watcher
	if err := fw.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start file watcher: %v", err)
	}

	h := &TaskHandle{
		taskConfig: &taskConfig,
		watcher:    fw,
		exitResult: &drivers.ExitResult{},
		startedAt:  time.Now(),
	}

	d.tasks[cfg.ID] = h

	driverHandle := drivers.NewTaskHandle(drivers.Version, cfg.ID, pluginName, cfg.AllocID)
	return driverHandle, nil, nil
}

func (d *Driver) RecoverTask(handle *drivers.TaskHandle) error {
	return nil
}

func (d *Driver) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	d.lock.RLock()
	handle, exists := d.tasks[taskID]
	d.lock.RUnlock()

	if !exists {
		return nil, drivers.ErrTaskNotFound
	}

	ch := make(chan *drivers.ExitResult)
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return ch, nil
}

func (d *Driver) StopTask(taskID string, timeout time.Duration, signal string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	handle, exists := d.tasks[taskID]
	if !exists {
		return drivers.ErrTaskNotFound
	}

	if handle.watcher != nil {
		handle.watcher.Stop()
	}

	return nil
}

func (d *Driver) DestroyTask(taskID string, force bool) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	handle, exists := d.tasks[taskID]
	if !exists {
		return drivers.ErrTaskNotFound
	}

	if handle.watcher != nil {
		handle.watcher.Stop()
	}

	delete(d.tasks, taskID)
	return nil
}

func (d *Driver) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	handle, exists := d.tasks[taskID]
	if !exists {
		return nil, drivers.ErrTaskNotFound
	}

	return &drivers.TaskStatus{
		ID:          taskID,
		Name:        handle.taskConfig.Paths[0],
		State:       drivers.TaskStateRunning,
		StartedAt:   handle.startedAt,
		CompletedAt: handle.completedAt,
		ExitResult:  handle.exitResult,
	}, nil
}

func (d *Driver) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	return nil, nil
}

func (d *Driver) validateConfig(config *TaskConfig) error {
	if len(config.Paths) == 0 {
		return fmt.Errorf("at least one path must be specified")
	}

	if len(config.Events) == 0 {
		return fmt.Errorf("at least one event type must be specified")
	}

	return nil
}
