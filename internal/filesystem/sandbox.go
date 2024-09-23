package filesystem

import (
	"fmt"
	"os"
	idpath "path"
	"path/filepath"
	"strings"
	"unicode"
)

type SandboxFS struct {
	// absolute slash path on host machine
	root string
	// absolute slash path of sandbox
	current string
}

func genTestFilename(str string) string {
	flip := true
	return strings.Map(func(r rune) rune {
		if flip {
			if unicode.IsLower(r) {
				u := unicode.ToUpper(r)
				if unicode.ToLower(u) == r {
					r = u
					flip = false
				}
			} else if unicode.IsUpper(r) {
				l := unicode.ToLower(r)
				if unicode.ToUpper(l) == r {
					r = l
					flip = false
				}
			}
		}
		return r
	}, str)
}

func CheckFileSystemCaseSensitive() bool {
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		panic(fmt.Errorf("could not determine the case-sensitivity of the filesystem, unable to create the temp directory, %w", err))
	}
	defer os.Remove(dir)
	alt := filepath.Join(filepath.Dir(dir), genTestFilename(filepath.Base(dir)))

	dInfo, err := os.Stat(dir)
	if err != nil {
		panic(fmt.Errorf("could not determine the case-sensitivity of the filesystem, %w", err))
	}

	aInfo, err := os.Stat(alt)
	if err != nil {
		// If the file doesn't exists, assume we are on a case-sensitive filesystem.
		if os.IsNotExist(err) {
			return true
		}

		panic(fmt.Errorf("could not determine the case-sensitivity of the filesystem, %w", err))
	}

	return !os.SameFile(dInfo, aInfo)
}

var FSCaseSensitive = CheckFileSystemCaseSensitive()

func (s *SandboxFS) resolveOsPath(path string) (string, error) {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "/") {
		path = cleanHeadSlash(path)
		path = idpath.Join(s.root, path)
	} else {
		current, _ := s.resolveOsPath(s.current)
		path = idpath.Join(current, path)
		if strings.HasPrefix(s.root, path) {
			return "", fmt.Errorf("the input path \"%s\" is out of sand box", path)
		}
	}
	return path, nil
}

func (s *SandboxFS) IsCaseSensitive() bool {
	return FSCaseSensitive
}

func (s *SandboxFS) Delete(path string) error {
	hostPath, err := s.resolveOsPath(path)
	if err != nil {
		return fmt.Errorf("unable to delete \"%s\", %w", path, err)
	}
	err = os.Remove(filepath.FromSlash(hostPath))
	if err != nil {
		return fmt.Errorf("unable to delete \"%s\", %w", path, err)
	}
	return nil
}