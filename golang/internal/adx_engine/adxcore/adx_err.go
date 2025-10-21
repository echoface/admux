package adxcore

var (
	// Predefined errors
	ErrMissingSSID = NewAdxError(1000, "missing ssid parameter")
)

type (
	// AdxError represents an ADX specific error
	AdxError struct {
		Code    int64
		Message string
	}
)

func NewAdxError(code int64, message string) *AdxError {
	return &AdxError{
		Message: message,
		Code:    code,
	}
}

// Error method for AdxError
func (e *AdxError) Error() string {
	return e.Message
}
