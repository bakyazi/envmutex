package service

import "github.com/bakyazi/envmutex/model"

type Service interface {
	GetEnvironments() ([]model.Environment, error)
	LockEnvironment(name, owner string) error
	ReleaseEnvironment(name, owner string) error
}
