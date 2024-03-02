package sheets

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	apierrors "github.com/bakyazi/envmutex/errors"
	"github.com/bakyazi/envmutex/model"
	"github.com/bakyazi/envmutex/sliceutil"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Service struct {
	service *sheets.Service
	sheetId string
}

func NewService(sheetId string, opts ...option.ClientOption) (*Service, error) {
	service, err := sheets.NewService(context.Background(), opts...)
	if err != nil {
		return nil, apierrors.NewAPIError(http.StatusInternalServerError, err)
	}
	return &Service{service: service, sheetId: sheetId}, nil

}

func (s *Service) GetEnvironments() ([]model.Environment, error) {
	resp, err := s.service.Spreadsheets.Values.Get(s.sheetId, "A2:E").Do()
	if err != nil {
		return nil, apierrors.NewAPIError(http.StatusInternalServerError, err)
	}

	return sliceutil.Map(resp.Values, func(t []interface{}, i int) model.Environment {
		date, _ := time.Parse(time.RFC850, t[3].(string))
		return model.Environment{
			Name:   t[0].(string),
			Status: t[1].(string),
			Owner:  t[2].(string),
			Date:   date,
		}
	}), nil
}

func (s *Service) LockEnvironment(name, owner string) error {
	envs, apiErr := s.GetEnvironments()
	if apiErr != nil {
		return apiErr
	}

	var index = -1
	var e model.Environment
	for i, env := range envs {
		if env.Name != name {
			continue
		}
		index = i
		e = env
		break
	}

	if index == -1 {
		return apierrors.NewAPIError(http.StatusNotFound, errors.New("environment not found"))
	}

	if e.Status != "Free" {
		return apierrors.NewAPIError(http.StatusLocked, errors.New("environment already locked"))
	}

	e.Status = "Locked"
	e.Owner = owner
	e.Date = time.Now()

	valRange := fmt.Sprintf("B%d:D%d", index+2, index+2)
	_, err := s.service.Spreadsheets.Values.Update(s.sheetId, valRange,
		&sheets.ValueRange{
			Range: valRange,
			Values: [][]any{
				{e.Status, e.Owner, e.Date.Format(time.RFC850)},
			},
		}).Do(googleapi.QueryParameter("valueInputOption", "RAW"))
	if err != nil {
		return apierrors.NewAPIError(http.StatusInternalServerError, err)
	}
	return nil
}

func (s *Service) ReleaseEnvironment(name, owner string) error {
	envs, apiErr := s.GetEnvironments()
	if apiErr != nil {
		return apiErr
	}

	var index = -1
	var e model.Environment
	for i, env := range envs {
		if env.Name != name {
			continue
		}
		index = i
		e = env
	}

	if index == -1 {
		return apierrors.NewAPIError(http.StatusNotFound, errors.New("environment not found"))
	}

	if e.Status != "Locked" {
		return apierrors.NewAPIError(http.StatusLocked, errors.New("environment already locked"))
	}

	if e.Owner != owner {
		return apierrors.NewAPIError(http.StatusForbidden, errors.New("not owned by user"))
	}
	e.Status = "Free"
	e.Owner = ""
	e.Date = time.Now()

	valRange := fmt.Sprintf("B%d:D%d", index+2, index+2)
	_, err := s.service.Spreadsheets.Values.Update(s.sheetId, valRange,
		&sheets.ValueRange{
			Range: valRange,
			Values: [][]any{
				{e.Status, e.Owner, e.Date.Format(time.RFC850)},
			},
		}).Do(googleapi.QueryParameter("valueInputOption", "RAW"))
	if err != nil {
		return apierrors.NewAPIError(http.StatusInternalServerError, err)
	}
	return nil
}
