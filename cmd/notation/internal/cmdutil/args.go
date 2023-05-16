package cmdutil

import (
	"errors"
	"fmt"
)

// ValidateArgsCount validates the arguments are of expected length.
func ValidateArgsCount(args []string, expLen int, missingErrMsg string) error {
	argsLength := len(args)
	if argsLength < expLen {
		return errors.New(missingErrMsg)
	} else if argsLength > expLen {
		return fmt.Errorf("accepts %d arg(s), received %d", expLen, argsLength)
	}
	return nil
}
