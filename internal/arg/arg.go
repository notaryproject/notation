package arg

import (
	"errors"
	"fmt"
)

// ValidateCount validates the arguments are of expected length.
func ValidateCount(args []string, expLen int, missingErrMsg string) error {
	argsLength := len(args)
	if argsLength < expLen {
		return errors.New(missingErrMsg)
	} else if argsLength > expLen {
		return fmt.Errorf("expected %d argument(s) but found %d, arguments: %v", expLen, argsLength, args)
	}
	return nil
}
