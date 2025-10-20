package usersession

import (
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

var UserSessionActorAddress = actor.NewAddress("locale", "session-actor")

func NewUserSessionActor() (*actor.Actor, error) {
	state, err := NewUserSessionActorState()
	if err != nil {
		return nil, err
	}
	a, err := actor.NewActor(UserSessionActorAddress, state)
	return &a, err
}

type UserSessionActorState struct {
	repository *sqlx.DB
}

func NewUserSessionActorState() (*UserSessionActorState, error) {
	repo, err := postgres.NewPostgresqlRepository()
	if err != nil {
		return nil, err
	}
	return &UserSessionActorState{
		repository: repo,
	}, nil
}

type CreateSessionItemMsgBody struct {
	SubjectID    string
	RefreshToken string
}

type CreateSessionItemMsgBodyResult struct {
	SessionID string
}

type RetriveSessionMsgBody struct {
	SessionId string
}

type RetriveSessionMsgBodyResult struct {
	Session UserSession
}

func (state *UserSessionActorState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case CreateSessionItemMsgBody:
		sid, err := state.createSession(payload.SubjectID, payload.RefreshToken)
		sids := ""
		if sid != nil {
			sids = sid.String()
		}
		if msg.WithReturn {
			returnMsg := actor.NewReturnMessage(CreateSessionItemMsgBodyResult{SessionID: sids}, msg)
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}

	case RetriveSessionMsgBody:
		session, err := state.getSession(payload.SessionId)
		if msg.WithReturn {
			returnMsg := actor.NewReturnMessage(RetriveSessionMsgBodyResult{session}, msg)
			msg.ReturnChan <- actor.NewWrappedMessage(&returnMsg, err)
		}
	}
}

var insertSessionSql = `-- insert session
		INSERT INTO locale.session (
			subject_id,
			refresh_token,
			refresh_counter,
			session_id,
			session_expire_at,
			archived
			)
		VALUES(
			:subject_id,
			:refresh_token,
			:refresh_counter,
			:session_id,
			:session_expire_at,
			:archived
		)
	`

// ON CONFLICT (subject_id)
// DO UPDATE SET
//
//	:refresh_token,
//	:refresh_counter,
//	:session_expire_at,
//	:archived
func (state UserSessionActorState) createSession(subjectId string, refreshToken string) (*uuid.UUID, error) {
	s, err := NewUserSession(subjectId, refreshToken)
	if err != nil {
		return nil, err
	}

	result, err := state.repository.NamedExec(insertSessionSql, s)
	if err != nil {
		return nil, err
	}

	numRows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if numRows != 1 {
		return nil, errors.New("unexpected number of rows affected")
	}

	return &s.SessionId, nil
}

func (state UserSessionActorState) getSession(sessionId string) (UserSession, error) {
	result := UserSession{}
	err := state.repository.Get(&result, "SELECT * FROM locale.session WHERE session_id = $1", sessionId)
	if err != nil {
		return UserSession{}, err
	}
	return result, nil
}

func (state UserSessionActorState) GetState() any {
	return nil
}

func (state UserSessionActorState) Shutdown() {
	err := state.repository.Close()
	if err != nil {
		slog.Error("error closing repository for user session actor", slog.String("error", err.Error()))
	}
	state.repository = nil
}
