package user

import (
	"errors"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

type UserActorState struct {
	repository *sqlx.DB
}

func newUserActorState() (*UserActorState, error) {
	repo, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}
	return &UserActorState{
		repository: repo,
	}, nil
}

var UserActorAddress = actor.NewAddress("locale", "user-actor")

func NewUserActor() (*actor.Actor, error) {
	state, err := newUserActorState()
	if err != nil {
		return nil, err
	}
	a, err := actor.NewActor(UserActorAddress, state)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

var insertUpdateUser string = `--insert sql
INSERT INTO
locale."user" (
	subject_id,
	email,
	name,
	role,
	picture,
	refresh_token
)
VALUES (
	:subject_id,
	:email,
	:name,
	:role,
	:picture,
	:refresh_token
)
ON CONFLICT (subject_id)
DO UPDATE SET
    email = :email,
    name = :name,
    role = :role,
    picture = :picture,
    refresh_token = :refresh_token;
`

func (state UserActorState) updateUser(user User) error {
	result, err := state.repository.NamedExec(insertUpdateUser, user)
	if err != nil {
		return err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if numRows != 1 {
		return errors.New("unexpected number of rows affected")
	}

	return nil
}

func (state UserActorState) getUser(subjectId string) (User, error) {
	var result User
	err := state.repository.Get(&result, "SELECT * FROM locale.user WHERE subject_id = $1;", subjectId)
	if err != nil {
		return User{}, err
	}
	return result, nil
}

type UpdateUserMessageBody struct {
	User UserInfoReponseBody
}

type UpdateUserMessageBodyResult struct {
	Success bool
}

type RetriveUserMessageBody struct {
	SubjectID string
}

type RetriveUserMessageBodyResult struct {
	User User
}

func (state *UserActorState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case UpdateUserMessageBody:
		u, err := NewUser(
			payload.User.SubjectID,
			payload.User.Email,
			payload.User.Name,
			payload.User.Picture,
			Translator, // TODO: make sense role hardcoded?
			payload.User.RefreshToken,
		)

		if err == nil {
			err = state.updateUser(u)
		}

		if msg.WithReturn {
			returnMsg := actor.NewReturnMessage(UpdateUserMessageBodyResult{err != nil}, msg)
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}

	case RetriveUserMessageBody:
		user, err := state.getUser(payload.SubjectID)
		if msg.WithReturn {
			returnMsg := actor.NewReturnMessage(RetriveUserMessageBodyResult{user}, msg)
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}
	}
}

func (state *UserActorState) GetState() any {
	return nil
}

func (state *UserActorState) Shutdown() {
	err := state.repository.Close()
	if err != nil {
		slog.Error("error closing database connection", slog.String("err", err.Error()))
	}
	state.repository = nil
}
