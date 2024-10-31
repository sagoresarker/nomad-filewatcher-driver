package watcher

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventType string

const (
	EventCreate EventType = "create"
	EventModify EventType = "modify"
	EventDelete EventType = "delete"
	EventRename EventType = "rename"
	EventChmod  EventType = "chmod"
)

type Event struct {
	Type      EventType `json:"type"`
	Path      string    `json:"path"`
	Timestamp time.Time `json:"timestamp"`
}

type EventHandler struct {
	Events chan Event
}

func NewEventHandler() *EventHandler {
	return &EventHandler{
		Events: make(chan Event, 100), // Buffer size of 100 events
	}
}

func (e *EventHandler) HandleEvent(event Event) error {
	select {
	case e.Events <- event:
		return nil
	default:
		return fmt.Errorf("event buffer full")
	}
}

func (e Event) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}

func IsValidEventType(eventType string) bool {
	switch EventType(eventType) {
	case EventCreate, EventModify, EventDelete, EventRename, EventChmod:
		return true
	default:
		return false
	}
}
