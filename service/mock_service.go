package service

import (
	"errors"
	"github.com/bakyazi/envmutex/model"
	"sort"
	"time"
)

type MockService struct {
	m map[string]model.Environment
}

func NewMockService(m map[string]model.Environment) *MockService {
	return &MockService{
		m: m,
	}
}

func (m *MockService) GetEnvironments() ([]model.Environment, error) {
	var result []model.Environment
	for _, v := range m.m {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result, nil
}

func (m *MockService) LockEnvironment(name, owner string) error {
	val, ok := m.m[name]
	if !ok {
		return errors.New("not found env")
	}

	if val.Owner != "" {
		return errors.New("already locked")
	}

	val.Owner = owner
	val.Status = "Locked"
	val.Date = time.Now()
	m.m[name] = val
	return nil
}

func (m *MockService) ReleaseEnvironment(name, owner string) error {
	val, ok := m.m[name]
	if !ok {
		return errors.New("not found env")
	}

	if val.Owner == "" {
		return errors.New("already released")
	}

	if val.Owner != owner {
		return errors.New("not permitted")
	}

	val.Owner = ""
	val.Status = "Free"
	val.Date = time.Now()
	m.m[name] = val
	return nil
}
