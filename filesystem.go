package v8tsgo

import (
	"container/list"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	v8 "rogchap.com/v8go"
)

type MemoryFS struct {
	Root          *MemoryDirNode
	current       *MemoryDirNode
	caseSensitive bool
}

type MemoryDirNode struct {
	parent   *MemoryDirNode
	name     string
	children list.List
	files    list.List
}

func (d *MemoryDirNode) Delete() {
	if d.parent == nil {
		return
	}
	childNode := d.parent.children.Front()
	for {
		if childNode != nil {
			child := childNode.Value.(*MemoryDirNode)
			if child == d {
				d.parent.children.Remove(childNode)
				return
			}
			childNode = childNode.Next()
		} else {
			break
		}
	}
}

func (d *MemoryDirNode) Clean() {
	d.children.Init()
	d.files.Init()
}

type MemoryFileNode struct {
	parent  *MemoryDirNode
	name    string
	content string
}

func (f *MemoryFileNode) Delete() {
	fileNode := f.parent.files.Front()
	for {
		if fileNode != nil {
			file := fileNode.Value.(*MemoryFileNode)
			if file == f {
				f.parent.files.Remove(fileNode)
				return
			}
			fileNode = fileNode.Next()
		} else {
			break
		}
	}
}

func NewFileNotExists(path string) error {
	return errors.New(fmt.Sprintf("file or dir %s not exists", path))
}

func (fs *MemoryFS) locate(path string) (*MemoryDirNode, *MemoryFileNode) {
	parts := strings.Split(path, "/")
	node := fs.current
	for i, part := range parts {
		if part != "" {
			found := false
			childElement := node.children.Front()
			for {
				if childElement != nil {
					child := childElement.Value.(*MemoryDirNode)
					if child.name == part {
						node = child
						found = true
						break
					}
					childElement = childElement.Next()
				} else {
					break
				}
			}
			if !found {
				fileNode := node.files.Front()
				for {
					if fileNode != nil {
						file := fileNode.Value.(*MemoryFileNode)
						if file.name == part && i == len(parts) - 1 {
							return file.parent, file
						}
						fileNode = fileNode.Next()
					} else {
						break
					}
				}
			}
			if !found {
				return nil, nil
			}
		}
	}
	return node, nil
}

func (fs *MemoryFS) Delete(path string) error {
	dir, file := fs.locate(path)
	if dir != nil {
		if file != nil {
			file.Delete()
		} else {
			if dir == fs.Root {
				dir.Clean()
			} else {
				dir.Delete()
			}
		}
		return nil
	} else {
		return NewFileNotExists(path)
	}
}

type V8FileSystem struct {
	// Gets if this file system is case sensitive.
	//   isCaseSensitive(): boolean
	fnIsCaseSensitive *v8.FunctionTemplate
	cbIsCaseSensitive func() (bool, error)

	// Asynchronously deletes the specified file or directory.
	//   delete(path: string): Promise<void>
	fnDelete *v8.FunctionTemplate
	// Synchronously deletes the specified file or directory.
	//   deleteSync(path: string): void
	fnDeleteSync *v8.FunctionTemplate
	cbDelete     func(path string) error

	// Reads all the child directories and files.
	// Implementers should have this return the full file path.
	//   readDirSync(dirPath: string): RuntimeDirEntry[]
	fnReadDirSync *v8.FunctionTemplate
	cbReadDir     func(dirPath string) ([]fs.FileInfo, error)

	// Asynchronously reads a file at the specified path.
	//	 readFile(filePath: string, encoding?: string): Promise<string>
	fnReadFile *v8.FunctionTemplate
	// Synchronously reads a file at the specified path.
	//	 readFileSync(filePath: string, encoding?: string): string
	fnReadFileSync *v8.FunctionTemplate
	cbReadFile     func(filePath string, encoding string) (string, error)

	// Asynchronously writes a file to the file system.
	//   writeFile(filePath: string, fileText: string): Promise<void>
	fnWriteFile *v8.FunctionTemplate
	// Synchronously writes a file to the file system.
	//   writeFileSync(filePath: string, fileText: string): void
	fnWriteFileSync *v8.FunctionTemplate
	cbWriteFile     func(filePath string, fileText string) error

	// Asynchronously creates a directory at the specified path.
	//   mkdir(dirPath: string): Promise<void>
	fnMkdir *v8.FunctionTemplate
	// Synchronously creates a directory at the specified path.
	//   mkdirSync(dirPath: string): void
	fnMkdirSync *v8.FunctionTemplate
	cbMkdir     func(dirPath string) error

	// Asynchronously moves a file or directory.
	//   move(srcPath: string, destPath: string): Promise<void>
	fnMove *v8.FunctionTemplate
	// Synchronously moves a file or directory.
	//   moveSync(srcPath: string, destPath: string): void
	fnMoveSync *v8.FunctionTemplate
	cbMove     func(srcPath string, destPath string) error

	// Asynchronously copies a file or directory.
	//   copy(srcPath: string, destPath: string): Promise<void>
	fnCopy *v8.FunctionTemplate
	// Synchronously copies a file or directory.
	//   copySync(srcPath: string, destPath: string): void
	fnCopySync *v8.FunctionTemplate
	cbCopy     func(srcPath string, destPath string) error

	// Asynchronously checks if a file exists.
	// Implementers should throw an `errors.FileNotFoundError` when it does not exist.
	//   fileExists(filePath: string): Promise<boolean>
	fnFileExists *v8.FunctionTemplate
	// Synchronously checks if a file exists.
	// Implementers should throw an `errors.FileNotFoundError` when it does not exist.
	//   fileExistsSync(filePath: string): boolean
	fnFileExistsSync *v8.FunctionTemplate
	cbFileExists     func(filePath string) (bool, error)

	// Asynchronously checks if a directory exists.
	//   directoryExists(dirPath: string): Promise<boolean>
	fnDirectoryExists *v8.FunctionTemplate
	// Synchronously checks if a directory exists.
	//   directoryExistsSync(dirPath: string): boolean
	fnDirectoryExistsSync *v8.FunctionTemplate
	cbDirectoryExists     func(dirPath string) (bool, error)

	// See https://nodejs.org/api/fs.html#fs_fs_realpathsync_path_options
	//   realpathSync(path: string): string
	fnRealpathSync *v8.FunctionTemplate
	cbRealpathSync func(path string) (string, error)

	// Gets the current directory of the environment.
	//   getCurrentDirectory(): string
	fnGetCurrentDirectory *v8.FunctionTemplate
	cbGetCurrentDirectory func() (string, error)

	// Uses pattern matching to find files or directories.
	//   glob(patterns: ReadonlyArray<string>): Promise<string[]>
	fnGlob *v8.FunctionTemplate
	// Synchronously uses pattern matching to find files or directories.
	//   globSync(patterns: ReadonlyArray<string>): string[]
	fnGlobSync *v8.FunctionTemplate
	cbGlob     func(patterns []string) ([]string, error)
}

func NewV8FileSystem(fs fs.FS, iso *v8.Isolate) *V8FileSystem {

	return nil
}

func (f *V8FileSystem) isCaseSensitive() bool {
	return true
}

func (f *V8FileSystem) delete(path string) error {
	return nil
}

func (f *V8FileSystem) readDir(dirPath string) ([]fs.File, error) {
	return nil, nil
}
