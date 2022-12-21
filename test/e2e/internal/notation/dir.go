package notation

import "path/filepath"

func NotationDir(userDir string) string {
	return filepath.Join(userDir, "notation")
}
