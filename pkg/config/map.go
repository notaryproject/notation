package config

// KeySuite is a named key suite with file paths.
type KeySuite struct {
	Name            string `json:"name"`
	KeyPath         string `json:"keyPath"`
	CertificatePath string `json:"certPath"`
}

// KeyMap is a set of KeySuite indexed by name.
// The overall performance is O(n) while the order of entries is persevered.
type KeyMap []KeySuite

// Append appends a uniquely named KeySuite to the map.
// Return true if new values are appended.
func (m *KeyMap) Append(name, keyPath, certPath string) bool {
	for _, ref := range *m {
		if ref.Name == name {
			return false
		}
	}
	*m = append(*m, KeySuite{
		Name:            name,
		KeyPath:         keyPath,
		CertificatePath: certPath,
	})
	return true
}

// Remove removes a named path from the map.
// Return true if an entry is found and removed.
func (m *KeyMap) Remove(name string) bool {
	for i, ref := range *m {
		if ref.Name == name {
			*m = append((*m)[:i], (*m)[i+1:]...)
			return true
		}
	}
	return false
}

// Get return the paths of the given name.
// Return true if found.
func (m KeyMap) Get(name string) (string, string, bool) {
	for _, ref := range m {
		if ref.Name == name {
			return ref.KeyPath, ref.CertificatePath, true
		}
	}
	return "", "", false
}

// CertificateReference is a named file path.
type CertificateReference struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// CertificateMap is a set of CertificateReference indexed by name.
// The overall performance is O(n) while the order of entries is persevered.
type CertificateMap []CertificateReference

// Append appends a uniquely named path to the map.
// Return true if new values are appended.
func (m *CertificateMap) Append(name, path string) bool {
	for _, ref := range *m {
		if ref.Name == name {
			return false
		}
	}
	*m = append(*m, CertificateReference{
		Name: name,
		Path: path,
	})
	return true
}

// Remove removes a named path from the map.
// Return true if an entry is found and removed.
func (m *CertificateMap) Remove(name string) bool {
	for i, ref := range *m {
		if ref.Name == name {
			*m = append((*m)[:i], (*m)[i+1:]...)
			return true
		}
	}
	return false
}

// Get return the path of the given name.
// Return true if found.
func (m CertificateMap) Get(name string) (string, bool) {
	for _, ref := range m {
		if ref.Name == name {
			return ref.Path, true
		}
	}
	return "", false
}

// PluginReference is a a named plugin.
type PluginReference struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// PluginMap is a set of PluginReference indexed by name.
// The overall performance is O(n) while the order of entries is persevered.
type PluginMap []PluginReference

// Append appends a uniquely named path to the map.
// Return true if new values are appended.
func (m *PluginMap) Append(name, path string) bool {
	for _, ref := range *m {
		if ref.Name == name {
			return false
		}
	}
	*m = append(*m, PluginReference{
		Name: name,
		Path: path,
	})
	return true
}

// Remove removes a named path from the map.
// Return true if an entry is found and removed.
func (m *PluginMap) Remove(name string) bool {
	for i, ref := range *m {
		if ref.Name == name {
			*m = append((*m)[:i], (*m)[i+1:]...)
			return true
		}
	}
	return false
}

// Get return the path of the given name.
// Return true if found.
func (m PluginMap) Get(name string) (string, bool) {
	for _, ref := range m {
		if ref.Name == name {
			return ref.Path, true
		}
	}
	return "", false
}
