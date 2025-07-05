package main

import (
	"log/slog"
	"os"

	"github.com/pix303/actor-lib/pkg/actor"
	"github.com/pix303/localemgmt-go/api/pkg/router"
)

func main() {
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	slog.SetDefault(logger)
	slog.Debug("logger setted")

	routerState, err := router.NewLocaleItemRouterState()
	if err != nil {
		slog.Error("error on startup router", slog.String("err", err.Error()))
		return
	}

	routerActor, err := actor.NewActor(
		actor.NewAddress("local", "router"),
		routerState,
	)
	if err != nil {
		slog.Error("error on startup router actor", slog.String("err", err.Error()))
		return
	}

	actor.RegisterActor(&routerActor)

	startEvent := router.StartRouter{}
	msg := actor.Message{
		From: actor.NewAddress("local", "main"),
		To:   actor.NewAddress("local", "router"),
		Body: startEvent,
	}
	actor.DispatchMessage(msg)

	ctx := actor.GetPostman().GetContext()
	<-ctx.Done()
	slog.Info("all shutdown... bye")
}
