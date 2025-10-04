package user

import (
	"errors"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/localemgmt-go/domain/pkg/user"
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
locale.user (
	subject_id,
	email,
	name,
	contexts,
	role
)
VALUES (
	:subject_id,
	:email,
	:name,
	:contexts,
	:role
)
ON CONFLICT (subject_id)
DO UPDATE SET
    email = :email,
    name = :name,
    contexts = :contexts,
    role = :role;
`

// refresh_token = :refresh_token,
// refresh_counter = :refresh_counter,
// session_id= :session_id,
// session_expire_at= :session_expire_at,

func (state UserActorState) updateUser(user User) error {
	udb := user.ConvertInUserForDB()
	result, err := state.repository.NamedExec(insertUpdateUser, udb)
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

var insertSession string = `--insert session sql
INSERT INTO
locale.session (
	subject_id,
	refresh_token,
	refresh_counter,
	session_id,
	session_expire_at
)
VALUES (
	:subject_id,
	:email,
	:name,
	:contexts,
	:role
)
ON CONFLICT (subject_id)
DO UPDATE SET
    email = :email,
    name = :name,
    contexts = :contexts,
    role = :role;
`

func (state UserActorState) createSession(usersession user.UserSession) error {
	result, err := state.repository.NamedExec(insertUpdateUser, udb)
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

func (state UserActorState) GetUser(subjectId string) (User, error) {
	var user User
	err := state.repository.Select(&user, "SELECT * FROM locale.user WHERE subject_id = $1;", subjectId)
	if err != nil {
		return user, err
	}
	return user, nil
}

type UpdateUserMessageBody struct {
	User UserInfoReponse
}

type RetriveUserMessageBody struct {
	SubjectID string
}

func (state *UserActorState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case UpdateUserMessageBody:
		u := NewUser(
			payload.User.SubjectID,
			payload.User.Email,
			payload.User.Name,
			payload.User.Picture,
			user.Translator,
		)

		err := state.updateUser(u)
		if msg.WithReturn {
			msg.ReturnChan <- actor.NewWrappedMessage(nil, err)
		}

	case RetriveUserMessageBody:
		user, err := state.GetUser(payload.SubjectID)
		returnMsg := actor.NewMessage(
			msg.From,
			msg.To,
			user,
			false,
		)
		if msg.WithReturn {
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}
	}
}

func (state *UserActorState) GetState() any {
	return nil
}

func (state *UserActorState) Shutdown() {
	err := state.repository.DB.Close()
	if err != nil {
		slog.Error("error closing database connection", slog.String("err", err.Error()))
	}
	state.repository = nil
}

// session user cases
// login
// - crate session record
// - return session id with coockies
// - middleware confirm session_id is not expired
// - if expired return 401
// - FE redirect to login
//
// refresh auth
// - middleware confirm session_id is not expired + check refresh counter limit
// - if refresh counter limit is reached: archive session + return 401
// - request new auth code to Google
// - update session_id record with: expred date + n day, new refresh token, +1 refresh counter
//
// logout
// - archive session_id
//
// revoke all sessions
// - archive all session
