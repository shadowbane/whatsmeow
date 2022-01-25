package queues

import (
	"container/list"
)

type Queue struct {
	Messages *list.List
}

func (q *Queue) Add(message interface{}) {
	q.Messages.PushBack(message)
}

func InitQueue() *Queue {
	return &Queue{
		Messages: list.New(),
	}
}
