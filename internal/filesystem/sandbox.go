package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
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

func (s *SandboxFS) ReadDir(dirPath string) ([]fs.FileInfo, error) {
	hostPath, err := s.resolveOsPath(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir \"%s\", %w", dirPath, err)
	}
	var result []fs.FileInfo
	root := filepath.FromSlash(hostPath)
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path != root {
				info, err := d.Info()
				if err != nil {
					if errors.Is(err, fs.ErrNotExist) {
						return nil
					} else {
						return err
					}
				}  else {
					result = append(result, info)
					return nil
				}
			} else {
				return nil
			}
		} else {
			return err
		}
	})
	if err != nil {
		err = fmt.Errorf("unable to read dir \"%s\", %w", dirPath, err)
	}
	return result, err
}

func (s *SandboxFS) ReadFile(filePath string, encoding string) (string, error) {
	hostPath, err := s.resolveOsPath(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file \"%s\", %w", filePath, err)
	}
	if strings.ToLower(encoding) != "utf8" && strings.ToLower(encoding) != "utf-8" {
		return "", fmt.Errorf("unable to read file \"%s\", only utf-8 is supported", filePath)
	}
	bytes, err := os.ReadFile(filepath.FromSlash(hostPath))
	if err != nil {
		return "", fmt.Errorf("unable to read file \"%s\", %w", filePath, err)
	}
	return string(bytes), nil
}

func (s *SandboxFS) WriteFile(filePath string, fileText string) error {
	hostPath, err := s.resolveOsPath(filePath)
	if err != nil {
		return fmt.Errorf("unable to write file \"%s\", %w", filePath, err)
	}
	err = os.WriteFile(filepath.FromSlash(hostPath), []byte(fileText), 0770)
	if err != nil {
		return fmt.Errorf("unable to write file \"%s\", %w", filePath, err)
	}
	return nil
}


func (s *SandboxFS) Mkdir(dirPath string) error {
	hostPath, err := s.resolveOsPath(dirPath)
	if err != nil {
		return fmt.Errorf("unable to make dir \"%s\", %w", dirPath, err)
	}
	err = os.MkdirAll(filepath.FromSlash(hostPath), 0770)
	if err != nil {
		return fmt.Errorf("unable to make dir \"%s\", %w", dirPath, err)
	}
	return nil
}

func (s *SandboxFS) Move(srcPath string, destPath string) error {
	hostSrcPath, err := s.resolveOsPath(srcPath)
	if err != nil {
		return fmt.Errorf("unable to move source path \"%s\" to dest path \"%s\", %w", srcPath, destPath, err)
	}
	hostDestPath, err := s.resolveOsPath(destPath)
	if err != nil {
		return fmt.Errorf("unable to move source path \"%s\" to dest path \"%s\", %w", srcPath, destPath, err)
	}
	osSrcPath := filepath.FromSlash(hostSrcPath)
	osDestPath := filepath.FromSlash(hostDestPath)
	_, err = os.Stat(osSrcPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("unable to move source path \"%s\" to dest path \"%s\", %w", srcPath, destPath, err)
	}
	err = os.Rename(osSrcPath, osDestPath)
	if err != nil {
		return fmt.Errorf("unable to move source path \"%s\" to dest path \"%s\", %w", srcPath, destPath, err)
	}
	return nil
}