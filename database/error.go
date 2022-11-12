package database

import "fmt"

// ErrInvalidOptionType is returned when an option is not of the expected type.
type ErrInvalidOptionType struct {
	gotT      interface{}
	expectedT interface{}
}

// Error implements error.Error interface.
func (e ErrInvalidOptionType) Error() string {
	return fmt.Sprintf("invalid option type %T (expected %T)", e.gotT, e.expectedT)
}
