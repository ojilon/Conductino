package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindFile walks upward from the anchor directory looking for a file
// by name. Returns the absolute path when found, error if it reaches
// the filesystem root without finding it.
//
// anchor: starting directory (pass os.Getwd() or dir of os.Executable())
// name:   filename to search for, e.g. "config.yaml"

func FindFile(anchor, name string) (string, error) {
	dir := anchor
	for {
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate), err == nil {
			return  candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			//reached the filesystem root
			return "",fmt.Errorf("file %d not found from %s upward", name, anchor)
		}
		dir = parent
	}

}