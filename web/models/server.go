package models

import (
	"time"
)

type (
	Server interface {
		Address() string
		OnlineAt() time.Time
		RunningTask() Task
	}

	server_item struct {
		address  string
		onlineat time.Time
		task     Task
	}
)

func (s *server_item) Address() string {
	return s.address
}

func (s *server_item) OnlineAt() time.Time {
	return s.onlineat
}

func (s *server_item) RunningTask() Task {
	return s.task
}
