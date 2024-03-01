package service

import "github.com/bakyazi/envmutex/model"

type Service interface {
	GetEnvironments() ([]model.Environment, error)
	LockEnvironment(name, owner string) error
	ReleaseEnvironment(name, owner string) error
}

type AuthService interface {
	Authenticate(name, password string) (string, error)
	ValidateUser(name, passwHash string) error
	ResetPassword(name, old, new string) error
}
