package mq

import (
	"errors"
)

type MessageQueue struct {
	Channel    chan string
	BufferSize int64
}

func Initialize(bufferSize int64) *MessageQueue {
	return &MessageQueue{
		Channel:    make(chan string, bufferSize),
		BufferSize: bufferSize,
	}
}

func (q *MessageQueue) BlockOrReadFront() string {
	return <-q.Channel
}

func (q *MessageQueue) WriteBack(msg string) error {
	select {
	case q.Channel <- msg:
		return nil
	default:
		return errors.New("failed to write to q. Possibly buffer is full")

	}
}

func (q *MessageQueue) BlockOrWriteBack(msg string) {
	q.Channel <- msg
}
