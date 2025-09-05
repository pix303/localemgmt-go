package main

import (
	"log/slog"
	"os"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/localemgmt-go/api/pkg/router"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/aggregate"
)

func main() {
	opts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))
	slog.SetDefault(logger)

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

	aggregateActor, err := aggregate.NewLocaleItemAggregateActor()
	if err != nil {
		slog.Error("error on startup aggregate actor", slog.String("err", err.Error()))
		return
	}
	actor.RegisterActor(aggregateActor)

	startEvent := router.StartRouter{}
	msg := actor.Message{
		From: actor.NewAddress("local", "main"),
		To:   actor.NewAddress("local", "router"),
		Body: startEvent,
	}
	actor.SendMessage(msg)

	ctx := actor.GetPostman().GetContext()
	<-ctx.Done()
	slog.Info("all shutdown... bye")
}
