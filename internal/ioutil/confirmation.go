package ioutil

import (
	"fmt"
	"io"
	"strings"
)

// AskForConfirmation prints a propmt to ask for confirmation before doing an
// action and takes user input as response.
func AskForConfirmation(r io.Reader, prompt string, confirmed bool) (bool, error) {
	if confirmed {
		return true, nil
	}

	fmt.Print(prompt, " [y/N] ")

	var response string
	if _, err := fmt.Fscanln(r, &response); err != nil {
		// in case of directly pressing Enter
		if err.Error() == "unexpected newline" {
			return false, nil
		}

		return false, err
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
