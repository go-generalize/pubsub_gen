package task

import (
	"time"
)

//go:generate pubsub_generator -pubsub Task

type Task struct {
	Desc    string
	Created time.Time
	Done    bool
	ID      int64
}
