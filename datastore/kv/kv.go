package kv

import (
	"errors"
	"fmt"
	"github.com/gr455/omnipresence/mq"
	"strings"
)

type KeyValueStore struct {
	Map   map[string]string
	ReadQ *mq.MessageQueue
}

func Initialize(readQ *mq.MessageQueue) *KeyValueStore {
	mp := &KeyValueStore{Map: make(map[string]string), ReadQ: readQ}
	err := mp.SubscribeToQueue()
	if err != nil {
		fmt.Printf("Error: cannot subscribe to queue: %v", error)
		return nil
	}

	return mp
}

func (mp *KeyValueStore) Get(key string) string {
	v, _ := mp.Map[key]
	return v
}

func (mp *KeyValueStore) Put(key, value string) {
	mp.Map[key] = value
}

func (mp *KeyValueStore) GetMap() {
	return mp.Map
}

func (mp *KeyValueStore) SubscribeToQueue() error {
	if mp.ReadQ == nil {
		return errors.New("no message queue set")
	}

	go func() {
		for msg := range mp.ReadQ {
			key, value, err := ParseMessage(msg)
			if err != nil {
				fmt.Printf("Error: cannot parse message: %v", error)
			}
		}
	}()

	return nil
}

func ParseMessage(msg string) (string, string, error) {
	words := strings.Fields(msg)
	if len(words) != 2 {
		return "", "", errors.New("illegal message format: %v", msg)
	}

	return words[0], words[1], nil
}
