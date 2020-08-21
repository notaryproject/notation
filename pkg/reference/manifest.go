package reference

import "time"

// Manifest to be signed
type Manifest struct {
	Descriptor
	Name       string    `json:"name,omitempty"`
	AccessedAt time.Time `json:"accessedAt,omitempty"`
}
