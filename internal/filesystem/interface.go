package filesystem

type FileSystem interface {
	IsCaseSensitive() bool
	Delete(path string) error
	ReadFile(filePath string, encoding string) (string, error)
	WriteFile(filePath string, fileText string) error
	Mkdir(dirPath string) error
	Move(srcPath string, destPath string) error
	Copy(srcPath string, destPath string) error
	FileExists(filePath string) (bool, error)
	DirectoryExists(dirPath string) (bool, error)
	Realpath(path string) (string, error)
	GetCurrentDirectory() (string, error)
	Glob(patterns []string) ([]string, error)
}


