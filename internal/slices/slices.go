package slices

type isser interface {
	Is(string) bool
}

// Index returns the index of the first occurrence of name in s,
// or -1 if not present.
func Index[E isser](s []E, name string) int {
	for i, v := range s {
		if v.Is(name) {
			return i
		}
	}
	return -1
}

// Contains reports whether name is present in s.
func Contains[E isser](s []E, name string) bool {
	return Index(s, name) >= 0
}

// Delete removes the elements s[i:i+1] from s,
// returning the modified slice.
func Delete[S ~[]E, E isser](s S, i int) S {
	return append(s[:i], s[i+1:]...)
}
