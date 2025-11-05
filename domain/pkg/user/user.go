package user

import (
	"errors"
	"os"

	"github.com/pix303/crypt-util-go/pkg/crypt"
)

var (
	ErrCryptKeyNotFound = errors.New("crypt key not found")
)

type UserInfoReponseBody struct {
	SubjectID    string `json:"sub"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Picture      string `json:"picture"`
	RefreshToken string
}

type UserRole int

const (
	Admin      UserRole = 1
	Translator UserRole = 2
	Reader     UserRole = 3
)

type User struct {
	SubjectID    string   `db:"subject_id" json:"subjectId"`
	Email        string   `db:"email" json:"email"`
	Name         string   `db:"name" json:"name"`
	Picture      string   `db:"picture" json:"picture"`
	Role         UserRole `db:"role" json:"role"`
	RefreshToken string   `db:"refresh_token" json:"-"`
	Contexts     []string `json:"contexts"`
}

func NewUser(sub string, email string, name string, picture string, userRole UserRole, refreshToken string) (User, error) {

	crykey := os.Getenv("REFRESH_TOKEN_CKEY")
	if crykey == "" {
		return User{}, ErrCryptKeyNotFound
	}

	crt, err := crypt.Encrypt([]byte(refreshToken), []byte(crykey))
	if err != nil {
		return User{}, err
	}

	return User{
		sub,
		email,
		name,
		picture,
		userRole,
		crt,
		make([]string, 0),
	}, nil
}
