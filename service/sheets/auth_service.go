package sheets

import (
	"errors"
	"fmt"
	apierrors "github.com/bakyazi/envmutex/errors"
	"github.com/bakyazi/envmutex/sliceutil"
	"github.com/golang-jwt/jwt"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
	"time"
)

type user struct {
	name     string
	password string
}

func (s *Service) Authenticate(name, password string) (string, error) {
	users, apiErr := s.getUsers()
	if apiErr != nil {
		return "", apiErr
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
				return "", apierrors.NewAPIError(http.StatusInternalServerError, err)
			}
			return tokenStr, nil
		}
		return "", apierrors.NewAPIError(http.StatusBadRequest, errors.New("wrong password"))

	}

	return "", apierrors.NewAPIError(http.StatusNotFound, errors.New("not found user"))
}

func (s *Service) getUsers() ([]user, error) {
	resp, err := s.service.Spreadsheets.Values.Get(s.sheetId, "Users!A2:C").Do()
	if err != nil {
		return nil, apierrors.NewAPIError(http.StatusInternalServerError, err)
	}

	return sliceutil.Map(resp.Values, func(t []interface{}, i int) user {
		return user{
			name:     t[0].(string),
			password: t[1].(string),
		}
	}), nil
}

func (s *Service) ValidateUser(name, password string) error {
	users, apiErr := s.getUsers()
	if apiErr != nil {
		return apiErr
	}

	for _, u := range users {
		if u.name == name {
			if u.password != password {
				return apierrors.NewAPIError(http.StatusUnauthorized, errors.New("not valid claim"))
			}
			return nil
		}
	}
	return apierrors.NewAPIError(http.StatusNotFound, errors.New("not found user"))
}

func (s *Service) ResetPassword(name, old, new string) error {
	users, apiErr := s.getUsers()
	if apiErr != nil {
		return apiErr
	}
	for i, u := range users {
		if name != u.name {
			continue
		}

		if u.password != old {
			return apierrors.NewAPIError(http.StatusBadRequest, errors.New("wrong old password"))
		}

		u.password = new
		return s.updatePassword(i+2, u)
	}
	return apierrors.NewAPIError(http.StatusNotFound, errors.New("not found user"))
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
	if err != nil {
		return apierrors.NewAPIError(http.StatusInternalServerError, err)
	}
	return nil
}
