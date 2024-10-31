package driver

import (
	"fmt"

	"github.com/hashicorp/nomad/plugins/shared/hclspec"
)

// FileWatcherConfig is the driver configuration
type FileWatcherConfig struct {
	Enabled         bool   `codec:"enabled"`
	StateDir        string `codec:"state_dir"`
	LogLevel        string `codec:"log_level"`
	MaxWatchPaths   int    `codec:"max_watch_paths"`
	EventBufferSize int    `codec:"event_buffer_size"`
}

// TaskConfig is the individual task configuration
type TaskConfig struct {
	Paths          []string          `codec:"paths"`           // Paths to watch
	Events         []string          `codec:"events"`          // Events to watch (create, modify, remove)
	ExecCommand    string            `codec:"exec_command"`    // Command to execute on events
	ExecArgs       []string          `codec:"exec_args"`       // Arguments for the command
	Environment    map[string]string `codec:"environment"`     // Environment variables
	RecursiveWatch bool              `codec:"recursive_watch"` // Watch subdirectories
	IgnorePatterns []string          `codec:"ignore_patterns"` // Patterns to ignore
	RetryInterval  int               `codec:"retry_interval"`  // Interval between retries in seconds
	MaxRetries     int               `codec:"max_retries"`     // Maximum number of retries
	Timeout        int               `codec:"timeout"`         // Timeout for command execution in seconds
}

// ConfigSpec is the specification of the plugin configuration
var configSpec = &hclspec.Spec{
	Block: &hclspec.Spec_Object{
		Object: &hclspec.Object{
			Attributes: map[string]*hclspec.Spec{
				"enabled": {
					Block: &hclspec.Spec_Bool{
						Bool: true,
					},
				},
				"state_dir": {
					Block: &hclspec.Spec_String{
						String_: &hclspec.String{
							Default: "/var/lib/nomad/filewatcher",
						},
					},
				},
				"log_level": {
					Block: &hclspec.Spec_String{
						String_: &hclspec.String{
							Default: "INFO",
						},
					},
				},
				"max_watch_paths": {
					Block: &hclspec.Spec_Number{
						Number: &hclspec.Number{
							Default: 100,
						},
					},
				},
				"event_buffer_size": {
					Block: &hclspec.Spec_Number{
						Number: &hclspec.Number{
							Default: 1000,
						},
					},
				},
			},
		},
	},
}

// TaskConfigSpec is the specification of task configuration
var taskConfigSpec = &hclspec.Spec{
	Block: &hclspec.Spec_Object{
		Object: &hclspec.Object{
			Attributes: map[string]*hclspec.Spec{
				"paths": {
					Block: &hclspec.Spec_Array{
						Array: &hclspec.Array{
							Values: []*hclspec.Spec{{
								Block: &hclspec.Spec_String{
									String_: &hclspec.String{},
								},
							}},
						},
					},
				},
				"events": {
					Block: &hclspec.Spec_Array{
						Array: &hclspec.Array{
							Values: []*hclspec.Spec{{
								Block: &hclspec.Spec_String{
									String_: &hclspec.String{},
								},
							}},
						},
					},
				},
				"exec_command": {
					Block: &hclspec.Spec_String{
						String_: &hclspec.String{},
					},
				},
				"exec_args": {
					Block: &hclspec.Spec_Array{
						Array: &hclspec.Array{
							Values: []*hclspec.Spec{{
								Block: &hclspec.Spec_String{
									String_: &hclspec.String{},
								},
							}},
						},
					},
				},
				"environment": {
					Block: &hclspec.Spec_Object{
						Object: &hclspec.Object{
							Attributes: map[string]*hclspec.Spec{},
						},
					},
				},
				"recursive_watch": {
					Block: &hclspec.Spec_Bool{
						Bool: false,
					},
				},
				"ignore_patterns": {
					Block: &hclspec.Spec_Array{
						Array: &hclspec.Array{
							Values: []*hclspec.Spec{{
								Block: &hclspec.Spec_String{
									String_: &hclspec.String{},
								},
							}},
						},
					},
				},
				"retry_interval": {
					Block: &hclspec.Spec_Number{
						Number: &hclspec.Number{
							Default: 30,
						},
					},
				},
				"max_retries": {
					Block: &hclspec.Spec_Number{
						Number: &hclspec.Number{
							Default: 3,
						},
					},
				},
				"timeout": {
					Block: &hclspec.Spec_Number{
						Number: &hclspec.Number{
							Default: 60,
						},
					},
				},
			},
		},
	},
}

// Validate validates the task configuration
func (tc *TaskConfig) Validate() error {
	if len(tc.Paths) == 0 {
		return fmt.Errorf("at least one path must be specified")
	}

	if len(tc.Events) == 0 {
		return fmt.Errorf("at least one event type must be specified")
	}

	if tc.ExecCommand == "" {
		return fmt.Errorf("exec_command must be specified")
	}

	// Validate event types
	validEvents := map[string]bool{
		"create": true,
		"modify": true,
		"remove": true,
		"rename": true,
		"chmod":  true,
	}

	for _, event := range tc.Events {
		if !validEvents[event] {
			return fmt.Errorf("invalid event type: %s", event)
		}
	}

	// Validate timeout
	if tc.Timeout < 0 {
		return fmt.Errorf("timeout must be non-negative")
	}

	// Validate retry settings
	if tc.RetryInterval < 0 {
		return fmt.Errorf("retry_interval must be non-negative")
	}

	if tc.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	return nil
}

// DefaultTaskConfig returns the default task configuration
func DefaultTaskConfig() *TaskConfig {
	return &TaskConfig{
		RecursiveWatch: false,
		RetryInterval:  30,
		MaxRetries:     3,
		Timeout:        60,
		Environment:    make(map[string]string),
	}
}

// Merge merges two TaskConfigs, with the other taking precedence
func (tc *TaskConfig) Merge(other *TaskConfig) *TaskConfig {
	result := *tc

	if len(other.Paths) > 0 {
		result.Paths = other.Paths
	}

	if len(other.Events) > 0 {
		result.Events = other.Events
	}

	if other.ExecCommand != "" {
		result.ExecCommand = other.ExecCommand
	}

	if len(other.ExecArgs) > 0 {
		result.ExecArgs = other.ExecArgs
	}

	if other.Environment != nil {
		if result.Environment == nil {
			result.Environment = make(map[string]string)
		}
		for k, v := range other.Environment {
			result.Environment[k] = v
		}
	}

	if other.RecursiveWatch {
		result.RecursiveWatch = true
	}

	if len(other.IgnorePatterns) > 0 {
		result.IgnorePatterns = other.IgnorePatterns
	}

	if other.RetryInterval > 0 {
		result.RetryInterval = other.RetryInterval
	}

	if other.MaxRetries > 0 {
		result.MaxRetries = other.MaxRetries
	}

	if other.Timeout > 0 {
		result.Timeout = other.Timeout
	}

	return &result
}
