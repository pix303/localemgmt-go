package main

import (
	"log/slog"
	"os"

	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/localemgmt-go/api/pkg/router"
	"github.com/pix303/localemgmt-go/domain/pkg/localeitem/aggregate"
	"github.com/pix303/localemgmt-go/domain/pkg/user"
	"github.com/pix303/localemgmt-go/domain/pkg/usersession"
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

	err = actor.RegisterActor(&routerActor)
	if err != nil {
		slog.Error("error on startup router actor", slog.String("err", err.Error()))
		return
	}

	aggregateActor, err := aggregate.NewLocaleItemAggregateActor()
	if err != nil {
		slog.Error("error on startup aggregate actor", slog.String("err", err.Error()))
		return
	}

	err = actor.RegisterActor(aggregateActor)
	if err != nil {
		slog.Error("error on startup router actor", slog.String("err", err.Error()))
		return
	}

	userActor, err := user.NewUserActor()
	if err != nil {
		slog.Error("error on startup user actor", slog.String("err", err.Error()))
		return
	}

	err = actor.RegisterActor(userActor)
	if err != nil {
		slog.Error("error on startup user actor", slog.String("err", err.Error()))
	}

	userSessionActor, err := usersession.NewUserSessionActor()
	if err != nil {
		slog.Error("error on startup user session actor", slog.String("err", err.Error()))
		return
	}

	err = actor.RegisterActor(userSessionActor)
	if err != nil {
		slog.Error("error on startup user session actor", slog.String("err", err.Error()))
	}

	startEvent := router.StartRouter{}
	msg := actor.Message{
		From: actor.NewAddress("local", "main"),
		To:   actor.NewAddress("local", "router"),
		Body: startEvent,
	}
	err = actor.SendMessage(msg)
	if err != nil {
		slog.Error("error on startup router", slog.String("err", err.Error()))
		return
	}

	ctx := actor.GetPostman().GetContext()
	<-ctx.Done()
	slog.Info("all shutdown... bye")
}
