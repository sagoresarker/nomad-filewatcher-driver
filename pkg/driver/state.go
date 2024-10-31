package driver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

type TaskState struct {
	ID          string      `json:"id"`
	StartedAt   time.Time   `json:"started_at"`
	CompletedAt time.Time   `json:"completed_at"`
	Config      *TaskConfig `json:"config"`
	Events      []string    `json:"events"`
	Paths       []string    `json:"paths"`
	ExitCode    int         `json:"exit_code"`
	Error       string      `json:"error,omitempty"`
	Status      string      `json:"status"`
}

type DriverState struct {
	Tasks map[string]*TaskState `json:"tasks"`
	lock  sync.RWMutex
}

func NewDriverState() *DriverState {
	return &DriverState{
		Tasks: make(map[string]*TaskState),
	}
}

func (s *DriverState) PutTask(id string, state *TaskState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Tasks[id] = state
	return s.persist()
}

func (s *DriverState) GetTask(id string) (*TaskState, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if state, exists := s.Tasks[id]; exists {
		return state, nil
	}
	return nil, fmt.Errorf("task not found: %s", id)
}

func (s *DriverState) DeleteTask(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Tasks, id)
	return s.persist()
}

func (s *DriverState) ListTasks() []*TaskState {
	s.lock.RLock()
	defer s.lock.RUnlock()

	tasks := make([]*TaskState, 0, len(s.Tasks))
	for _, task := range s.Tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

func (s *DriverState) persist() error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %v", err)
	}

	statePath := filepath.Join("/var/lib/nomad/filewatcher", "state.json")
	if err := ioutil.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %v", err)
	}

	return nil
}

func (s *DriverState) Restore() error {
	statePath := filepath.Join("/var/lib/nomad/filewatcher", "state.json")
	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("failed to read state file: %v", err)
	}

	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("failed to unmarshal state: %v", err)
	}

	return nil
}

func (s *DriverState) UpdateTaskStatus(id string, status string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if task, exists := s.Tasks[id]; exists {
		task.Status = status
		return s.persist()
	}
	return fmt.Errorf("task not found: %s", id)
}

func (s *DriverState) RecordTaskCompletion(id string, exitCode int, err error) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if task, exists := s.Tasks[id]; exists {
		task.CompletedAt = time.Now()
		task.ExitCode = exitCode
		if err != nil {
			task.Error = err.Error()
		}
		task.Status = "completed"
		return s.persist()
	}
	return fmt.Errorf("task not found: %s", id)
}
