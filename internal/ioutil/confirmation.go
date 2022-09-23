package ioutil

import (
	"bufio"
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

	scanner := bufio.NewScanner(r)
	if ok := scanner.Scan(); !ok {
		return false, scanner.Err()
	}
	response := scanner.Text()

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	default:
		fmt.Println("Operation cancelled.")
		return false, nil
	}
}
