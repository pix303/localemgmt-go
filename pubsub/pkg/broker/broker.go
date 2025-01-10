package broker

import (
	"sync"
)

type Broker struct {
	topicChannels map[string][]chan string
	closed        bool
	guard         sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		make(map[string][]chan string),
		false,
		sync.Mutex{},
	}
}

func (b *Broker) Subscribe(topic string) chan string {
	b.guard.Lock()
	defer b.guard.Unlock()

	var c = make(chan string)
	channels := b.topicChannels[topic]
	channels = append(channels, c)
	b.topicChannels[topic] = channels
	return c
}

func (b *Broker) Unsubscribe(topic string) {
	b.guard.Lock()
	defer b.guard.Unlock()

	for _, c := range b.topicChannels[topic] {
		close(c)
	}

	delete(b.topicChannels, topic)
}

func (b *Broker) Publish(topic, message string) {
	b.guard.Lock()
	defer b.guard.Unlock()

	channels := b.topicChannels[topic]
	for _, c := range channels {
		c <- message
	}
}

func (b *Broker) Close() {
	if b.closed == true {
		return
	}

	for k := range b.topicChannels {
		b.Unsubscribe(k)
	}
}
