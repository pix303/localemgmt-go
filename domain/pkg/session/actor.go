package usersession

import (
	"github.com/jmoiron/sqlx"
	"github.com/pix303/cinecity/pkg/actor"
	"github.com/pix303/postgres-util-go/pkg/postgres"
)

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

type UserSessionActorState struct {
	repository *sqlx.DB
}

var UserSessionActorAddress = actor.NewAddress("locale", "session-actor")

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

func (state *UserSessionActorState) Process(msg actor.Message) {
	switch payload := msg.Body.(type) {
	case CreateSessionItemMsgBody:

	}
}


var insertSession = `
		INSERT INTO locale.sessions ()
	`
func (state UserActorState) createSession(subjectId string, refreshToken string) error {
		s := NewUserSession(subjectID, refreshToken)
	result, err := state.repository.NamedExec(, usersession)
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
