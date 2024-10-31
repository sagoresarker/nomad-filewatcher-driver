package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/go-hclog"
)

type FileWatcher struct {
	watcher        *fsnotify.Watcher
	logger         hclog.Logger
	paths          []string
	events         []string
	execCommand    string
	execArgs       []string
	environment    map[string]string
	ignorePatterns []string
	recursiveWatch bool
	stopCh         chan struct{}
}

func NewFileWatcher(
	logger hclog.Logger,
	paths []string,
	events []string,
	execCommand string,
	execArgs []string,
	environment map[string]string,
	ignorePatterns []string,
	recursiveWatch bool,
) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	return &FileWatcher{
		watcher:        watcher,
		logger:         logger,
		paths:          paths,
		events:         events,
		execCommand:    execCommand,
		execArgs:       execArgs,
		environment:    environment,
		ignorePatterns: ignorePatterns,
		recursiveWatch: recursiveWatch,
		stopCh:         make(chan struct{}),
	}, nil
}

func (fw *FileWatcher) Start() error {
	// Add paths to watch
	for _, path := range fw.paths {
		if fw.recursiveWatch {
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return fw.watcher.Add(path)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("failed to add path recursively: %v", err)
			}
		} else {
			if err := fw.watcher.Add(path); err != nil {
				return fmt.Errorf("failed to add path: %v", err)
			}
		}
	}

	// Start watching
	go fw.watch()
	return nil
}

func (fw *FileWatcher) watch() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if fw.shouldHandle(event) {
				fw.handleEvent(event)
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			fw.logger.Error("watcher error", "error", err)
		case <-fw.stopCh:
			return
		}
	}
}

func (fw *FileWatcher) shouldHandle(event fsnotify.Event) bool {
	// Check if event type is in our list
	eventType := eventToString(event)
	if !containsString(fw.events, eventType) {
		return false
	}

	// Check ignore patterns
	for _, pattern := range fw.ignorePatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(event.Name)); matched {
			return false
		}
	}

	return true
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	fw.logger.Info("file event detected",
		"path", event.Name,
		"operation", event.Op.String(),
	)

	if fw.execCommand == "" {
		return
	}

	// Execute command
	cmd := exec.Command(fw.execCommand, fw.execArgs...)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("WATCHER_EVENT_PATH=%s", event.Name),
		fmt.Sprintf("WATCHER_EVENT_OP=%s", event.Op.String()),
	)

	// Add custom environment variables
	for k, v := range fw.environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fw.logger.Error("command execution failed",
			"error", err,
			"output", string(output),
		)
		return
	}

	fw.logger.Info("command executed successfully",
		"output", string(output),
	)
}

func (fw *FileWatcher) Stop() {
	close(fw.stopCh)
	fw.watcher.Close()
}

func eventToString(event fsnotify.Event) string {
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		return "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		return "modify"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	default:
		return "unknown"
	}
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
