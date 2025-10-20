package user

import (
	"github.com/lib/pq"
)

type UserInfoReponseBody struct {
	SubjectID     string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	RefreshToken  string
}

type UserRole int

const (
	Admin      UserRole = 1
	Translator UserRole = 2
	Reader     UserRole = 3
)

type UserBase struct {
	SubjectID string   `db:"subject_id" json:"subjectId"`
	Email     string   `db:"email" json:"email"`
	Name      string   `db:"name" json:"name"`
	Picture   string   `db:"picture" json:"picture"`
	Role      UserRole `db:"role" json:"role"`
}

type User struct {
	UserBase
	Contexts []string `json:"contexts"`
}

type UserForDB struct {
	UserBase
	Contexts pq.StringArray `db:"contexts"`
}

func NewUser(sub string, email string, name string, picture string, userRole UserRole) User {
	return User{
		UserBase: UserBase{
			sub,
			email,
			name,
			picture,
			userRole,
		},
		Contexts: make([]string, 0),
	}
}

func (u *User) ConvertInUserForDB() UserForDB {
	return UserForDB{
		UserBase: u.UserBase,
		Contexts: pq.StringArray(u.Contexts),
	}
}

func (u *UserForDB) ConvertInUser() User {
	return User{
		UserBase: u.UserBase,
		Contexts: []string(u.Contexts),
	}

}
