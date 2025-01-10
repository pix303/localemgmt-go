package main

import (
	"fmt"
	"time"

	"github.com/pix303/localemgmt-go/pubsub/pkg/broker"
)

func printMsg(c chan string) {
	for msg := range c {
		fmt.Printf("main --> %s\n", msg)
	}
}

func printMsg2(c chan string) {
	for msg := range c {
		fmt.Printf("secn --> %s\n", msg)
	}
}

func main() {
	b := broker.NewBroker()

	c := b.Subscribe("main")
	c2 := b.Subscribe("sec")

	go printMsg(c)
	go printMsg2(c2)

	b.Publish("main", "ciao")
	b.Publish("main", "ciao 2")
	b.Publish("main", "ciao 3")
	b.Publish("sec", "ciao sec 1")
	b.Publish("sec", "ciao sec 2")

	b.Unsubscribe("sec")
	b.Publish("main", "ciao 4")
	b.Publish("main", "ciao 5")

	b.Publish("sec", "ciao sec 3")
	b.Publish("sec", "ciao sec 4")

	time.Sleep(100 * time.Millisecond)
}
