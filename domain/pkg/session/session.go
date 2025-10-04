// session use cases
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

package usersession

import (
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	SubjectID      string    `db:"subject_id"`
	SessionId      uuid.UUID `db:"session_id"`
	ExpireAt       time.Time `db:"session_expire_at"`
	RefreshToken   string    `db:"refresh_token"`
	RefreshCounter int       `db:"refresh_counter"`
	Archived       bool      `db:"archived"`
}

func NewUserSession(subjectId string, refreshToken string) UserSession {
	expireAt := time.Now().AddDate(0, 0, 10)
	return UserSession{
		SubjectID:      subjectId,
		SessionId:      uuid.New(),
		ExpireAt:       expireAt,
		RefreshToken:   refreshToken,
		RefreshCounter: 0,
		Archived:       false,
	}
}
