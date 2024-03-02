package errors

func NewAPIError(status int, err error) error {
	return &APIError{StatusCode: status, Err: err}
}

type APIError struct {
	StatusCode int
	Err        error
}

func (a *APIError) Error() string {
	if a.Err != nil {
		return ""
	}
	return a.Err.Error()
}
