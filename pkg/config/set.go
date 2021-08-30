package config

// FileReference is a named file path.
type FileReference struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// FileSet is a set of FileReference indexed by name.
type FileSet []FileReference

// Append appends a uniquely named path to the set.
// Return true if new values are appended.
func (s *FileSet) Append(name, path string) bool {
	for _, ref := range *s {
		if ref.Name == name {
			return false
		}
	}
	*s = append(*s, FileReference{
		Name: name,
		Path: path,
	})
	return true
}

// Remove removes a named path from the set.
// Return true if an entry is found and removed.
func (s *FileSet) Remove(name string) bool {
	for i, ref := range *s {
		if ref.Name == name {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return true
		}
	}
	return false
}

// Get return the path of the given name.
// Return true if found.
func (s FileSet) Get(name string) (string, bool) {
	for _, ref := range s {
		if ref.Name == name {
			return ref.Path, true
		}
	}
	return "", false
}
