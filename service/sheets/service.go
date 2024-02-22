package sheets

import (
	"context"
	"errors"
	"fmt"
	"github.com/bakyazi/envmutex/model"
	"github.com/bakyazi/envmutex/sliceutil"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"time"
)

type Service struct {
	service *sheets.Service
	sheetId string
}

func NewService(sheetId string, opts ...option.ClientOption) (*Service, error) {
	service, err := sheets.NewService(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &Service{service: service, sheetId: sheetId}, nil

}

func (s Service) GetEnvironments() ([]model.Environment, error) {
	resp, err := s.service.Spreadsheets.Values.Get(s.sheetId, "A2:A").Do()
	if err != nil {
		return nil, err
	}

	return sliceutil.Map(resp.Values, func(t []interface{}, i int) model.Environment {
		return model.Environment{
			Name:   t[0].(string),
			Status: t[1].(string),
			Owner:  t[2].(string),
			Date:   t[3].(time.Time),
		}
	}), nil
}

func (s Service) LockEnvironment(name, owner string) error {
	envs, err := s.GetEnvironments()
	if err != nil {
		return err
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
		return errors.New("not found environment")
	}

	if e.Status != "Free" {
		return errors.New("environment already locked")
	}

	e.Status = "Locked"
	e.Owner = owner
	e.Date = time.Now()

	valRange := fmt.Sprintf("B%d:D%d", index, index)
	_, err = s.service.Spreadsheets.Values.Update(s.sheetId, valRange,
		&sheets.ValueRange{
			Range: valRange,
			Values: [][]any{
				{e.Status, e.Owner, e.Date},
			},
		}).Do(googleapi.QueryParameter("valueInputOption", "RAW"))
	return err
}

func (s Service) ReleaseEnvironment(name, owner string) error {
	envs, err := s.GetEnvironments()
	if err != nil {
		return err
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
		return errors.New("not found environment")
	}

	if e.Status != "Locked" {
		return errors.New("environment already released")
	}

	if e.Owner != owner {
		return errors.New("not owned by user")
	}
	e.Status = "Free"
	e.Owner = ""
	e.Date = time.Now()

	valRange := fmt.Sprintf("B%d:D%d", index, index)
	_, err = s.service.Spreadsheets.Values.Update(s.sheetId, valRange,
		&sheets.ValueRange{
			Range: valRange,
			Values: [][]any{
				{e.Status, e.Owner, e.Date},
			},
		}).Do(googleapi.QueryParameter("valueInputOption", "RAW"))
	return err
}
