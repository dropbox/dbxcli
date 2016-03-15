package apierror

type ApiError struct {
	ErrorSummary string `json:"error_summary"`
}

// implement the error interface
func (e ApiError) Error() string {
	return e.ErrorSummary
}
