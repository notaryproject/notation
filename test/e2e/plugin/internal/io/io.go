package io

import (
	"encoding/json"
	"os"
)

// UnmarshalRequest read the STDIN and unmarshal to struct pointed by req.
func UnmarshalRequest(req any) error {
	return json.NewDecoder(os.Stdin).Decode(req)
}

// PrintResponse marshal and print the struct data pointed by resp.
func PrintResponse(resp any) error {
	return json.NewEncoder(os.Stdout).Encode(resp)
}
