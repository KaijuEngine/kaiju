package concurrent

import (
	"sync"
)

type MessageQueue[T any] struct {
	mutex       sync.Mutex
	messages    []T
	flushBuffer []T
}

func NewMessageQueue[T any]() *MessageQueue[T] {
	return &MessageQueue[T]{
		messages: make([]T, 0),
	}
}

func (q *MessageQueue[T]) Enqueue(msg T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.messages = append(q.messages, msg)
}

func (q *MessageQueue[T]) Flush() []T {
	q.mutex.Lock()
	q.flushBuffer = q.flushBuffer[:0]
	q.messages, q.flushBuffer = q.flushBuffer, q.messages
	q.mutex.Unlock()
	return q.flushBuffer
}
