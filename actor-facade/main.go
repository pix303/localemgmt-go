package main

import (
	"github.com/pix303/actor-lib/pkg/actor"
	"github.com/pix303/localemgmt-go/api/pkg/router"
)

func main() {
	routerActor := actor.NewActor(
		actor.NewPID("local", "router"),
		router.NewLocaleItemRouterState(),
	)
	routerActor.Activate()
	ds := actor.NewActorDispatcher()
	ds.RegisterActor(&routerActor)
	var startEvent router.StartRouter = 0
	msg := actor.Message{
		From: *actor.NewPID("local", "main"),
		To:   *actor.NewPID("local", "router"),
		Body: startEvent,
	}

	ds.DispatchMessage(msg)

	select {}
}
