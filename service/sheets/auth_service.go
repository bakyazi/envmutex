package sheets

import (
	"context"
	"errors"
	"fmt"
	"github.com/bakyazi/envmutex/service"
	"github.com/bakyazi/envmutex/sliceutil"
	"github.com/golang-jwt/jwt"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"os"
	"time"
)

type user struct {
	name     string
	password string
}

func NewAuthService(sheetId string, opts ...option.ClientOption) (service.AuthService, error) {
	srv, err := sheets.NewService(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &Service{service: srv, sheetId: sheetId}, nil

}

func (s *Service) Authenticate(name, password string) (string, error) {
	users, err := s.getUsers()
	if err != nil {
		return "", err
	}
	for _, u := range users {
		if name != u.name {
			continue
		}

		if u.password == password {
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"name": name,
				"exp":  time.Now().Add(time.Hour).Unix(),
				"nbf":  time.Now().Unix(),
			})

			tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_PRIV_KEY")))
			if err != nil {
				return "", err
			}
			return tokenStr, nil
		}
		return "", errors.New("wrong password")

	}

	return "", errors.New("not found user")
}

func (s *Service) getUsers() ([]user, error) {
	resp, err := s.service.Spreadsheets.Values.Get(s.sheetId, "Users!A2:C").Do()
	if err != nil {
		return nil, err
	}

	return sliceutil.Map(resp.Values, func(t []interface{}, i int) user {
		return user{
			name:     t[0].(string),
			password: t[1].(string),
		}
	}), nil
}

func (s *Service) ValidateUser(name, passwHash string) error {
	users, err := s.getUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.name == name {
			if u.password != passwHash {
				return errors.New("not valid claim")
			}
			return nil
		}
	}
	return errors.New("not found user")
}

func (s *Service) ResetPassword(name, old, new string) error {
	users, err := s.getUsers()
	if err != nil {
		return err
	}
	for i, u := range users {
		if name != u.name {
			continue
		}

		if u.password != old {
			return errors.New("wrong old password")
		}

		u.password = new
		return s.updatePassword(i+2, u)
	}
	return errors.New("user not found")
}

func (s *Service) updatePassword(index int, u user) error {
	valRange := fmt.Sprintf("Users!B%d:B%d", index, index)
	_, err := s.service.Spreadsheets.Values.Update(s.sheetId, valRange,
		&sheets.ValueRange{
			Range: valRange,
			Values: [][]any{
				{u.password},
			},
		}).Do(googleapi.QueryParameter("valueInputOption", "RAW"))
	return err
}
