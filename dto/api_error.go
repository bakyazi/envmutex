package dto

type APIError struct {
	StatusCode int
	Err        error
}
