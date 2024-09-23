package filesystem

import (
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

type MemoryFS struct {
	root          *MemoryDirNode
	current       *MemoryDirNode
	caseSensitive bool
}

func NewMemoryFS(caseSensitive bool) *MemoryFS {
	root := &MemoryDirNode{
		modeTime: time.Now(),
	}
	return &MemoryFS{
		root: root,
		current: root,
		caseSensitive: caseSensitive,
	}
}

type MemoryDirNode struct {
	parent   *MemoryDirNode
	name     string
	children map[string]*MemoryDirNode
	files    map[string]*MemoryFileNode
	size     int64
	modeTime time.Time
}

func (d *MemoryDirNode) Delete() bool {
	if d.parent == nil {
		return false
	}
	me, ok := d.parent.children[d.name]
	if ok && me == d {
		delete(d.parent.children, d.name)
		d.parent.size -= d.Size()
		d.parent.modeTime = time.Now()
		return true
	} else {
		return false
	}
}

func (d *MemoryDirNode) Name() string {
	return d.name
}

func (d *MemoryDirNode) Size() int64 {
	return d.size
}

func (d *MemoryDirNode) Mode() fs.FileMode {
	return fs.ModePerm | fs.ModeDir
}

func (d *MemoryDirNode) ModTime() time.Time {
	return d.modeTime
}

func (d *MemoryDirNode) IsDir() bool {
	return true
}

func (d *MemoryDirNode) Sys() any {
	return nil
}

func (d *MemoryDirNode) Clean() {
	d.children = nil
	d.files = nil
	d.size = 0
	d.modeTime = time.Now()
}

func (d *MemoryDirNode) fullPath(sb *strings.Builder) {
	if d.parent != nil {
		d.parent.fullPath(sb)
	}
	sb.WriteString("/")
	sb.WriteString(d.name)
}

func (d *MemoryDirNode) FullPath() string {
	var sb strings.Builder
	d.fullPath(&sb)
	return sb.String()
}

func (d *MemoryDirNode) deepCopy(modTime time.Time) *MemoryDirNode {
	var children map[string]*MemoryDirNode
	if d.children != nil {
		children = make(map[string]*MemoryDirNode, len(d.children))
		for key, child := range d.children {
			children[key] = child.deepCopy(modTime)
		}	
	}
	var files map[string]*MemoryFileNode
	if d.files != nil {
		files = make(map[string]*MemoryFileNode, len(d.files))
		for key, file := range d.files {
			files[key] = file.copy(modTime)
		}	
	}
	return &MemoryDirNode{
		parent:   d.parent,
		name:     d.name,
		children: children,
		files:    files,
		size:     d.size,
		modeTime: modTime,
	}
}

func (d *MemoryDirNode) merge(dir *MemoryDirNode, overwrite bool) bool {
	if d.name != dir.name {
		return false
	}
	for dirName, dirChild := range dir.children {
		child, ok := d.children[dirName]
		if ok {
			child.merge(dirChild, overwrite)
		} else {
			d.children[dirName] = dirChild
			d.size += dirChild.Size()
			if dirChild.modeTime.After(d.modeTime) {
				d.modeTime = dirChild.modeTime
			}
		}
	}
	for fileName, dirFile := range dir.files {
		file, ok := d.files[fileName]
		if ok {
			if overwrite {
				file.overwrite(dirFile)
			}
		} else {
			d.files[fileName] = dirFile
			d.size += dirFile.Size()
			if dirFile.modeTime.After(d.modeTime) {
				d.modeTime = dirFile.modeTime
			}
		}
	}
	return true
}

type MemoryFileNode struct {
	parent   *MemoryDirNode
	name     string
	content  string
	modeTime time.Time
}

func (f *MemoryFileNode) Name() string {
	return f.name
}

func (f *MemoryFileNode) Size() int64 {
	return int64(len(f.content))
}

func (f *MemoryFileNode) Mode() fs.FileMode {
	return fs.ModePerm
}

func (f *MemoryFileNode) ModTime() time.Time {
	return f.modeTime
}

func (f *MemoryFileNode) IsDir() bool {
	return false
}

func (f *MemoryFileNode) Sys() any {
	return nil
}

func (f *MemoryFileNode) fullPath(sb *strings.Builder) {
	f.parent.fullPath(sb)
	sb.WriteString("/")
	sb.WriteString(f.name)
}

func (f *MemoryFileNode) FullPath() string {
	var sb strings.Builder
	f.fullPath(&sb)
	return sb.String()
}

func (f *MemoryFileNode) overwrite(file *MemoryFileNode) {
	sizeDiff := file.Size() - f.Size()
	f.content = file.content
	if file.modeTime.After(f.modeTime) {
		f.modeTime = file.modeTime
		if f.parent != nil {
			f.parent.modeTime = file.modeTime
		}
	}
	if f.parent != nil {
		f.parent.size += sizeDiff
	}

}

func (f *MemoryFileNode) Delete() bool {
	if f.parent == nil {
		return false
	}
	me, ok := f.parent.files[f.name]
	if ok && me == f {
		delete(f.parent.files, f.name)
		f.parent.size -= f.Size()
		f.parent.modeTime = time.Now()
		return true
	} else {
		return false
	}
}

func (f *MemoryFileNode) copy(modTime time.Time) *MemoryFileNode {
	return &MemoryFileNode{
		parent:   f.parent,
		name:     f.name,
		content:  f.content,
		modeTime: modTime,
	}
}

func NewFileOrDirNotExists(path string) error {
	return fmt.Errorf("%w, path: %s", fs.ErrNotExist, path)
}

func NewNotDir(path string) error {
	return fmt.Errorf("%w, the input path \"%s\" is not a dir", fs.ErrInvalid, path)
}

func NewNotFile(path string) error {
	return fmt.Errorf("%w, the input path \"%s\" is not a file", fs.ErrInvalid, path)
}

func (fs *MemoryFS) IsCaseSensitive() bool {
	return fs.caseSensitive
}

func (fs *MemoryFS) normName(name string) string {
	if fs.caseSensitive {
		return name
	} else {
		return strings.ToLower(name)
	}
}

func (fs *MemoryFS) locate(path string, dirOfFile bool) (*MemoryDirNode, *MemoryFileNode) {
	parts := strings.Split(path, "/")
	node := fs.current
	var child *MemoryDirNode
	var file *MemoryFileNode
	for i, part := range parts {
		if part != "" {
			found := false
			child, found = node.children[part]
			if !found {
				file, found = node.files[part]
				if found {
					if i == len(parts)-1 {
						return file.parent, file
					} else {
						return nil, nil
					}
				}
			}
			if !found {
				if dirOfFile && i == len(parts)-1 {
					return node, nil
				} else {
					return nil, nil
				}
			} else {
				node = child
			}
		}
	}
	return node, nil
}

func isFilePath(path string) bool {
	return path != "" && !strings.HasSuffix(path, "/")
}

func (fs *MemoryFS) resolve(path string) string {
	for {
		if strings.HasSuffix(path, "/") {
			path = path[0 : len(path)-1]
		} else {
			break
		}
	}
	path = fs.normName(path)
	if strings.HasPrefix(path, "/") {
		return path
	} else {
		currentPath := fs.current.FullPath()
		if currentPath == "/" {
			return currentPath + path
		} else {
			return currentPath + "/" + path
		}
	}
}

func (fs *MemoryFS) Delete(path string) error {
	path = fs.resolve(path)
	dir, file := fs.locate(path, false)
	if dir != nil {
		if file != nil {
			file.Delete()
		} else {
			if dir == fs.root {
				dir.Clean()
			} else {
				dir.Delete()
			}
		}
		return nil
	} else {
		return NewFileOrDirNotExists(path)
	}
}

func (fs *MemoryFS) ReadDir(dirPath string) ([]fs.FileInfo, error) {
	dirPath = fs.resolve(dirPath)
	dir, file := fs.locate(dirPath, false)
	if dir == nil {
		return nil, NewFileOrDirNotExists(dirPath)
	}
	if file != nil {
		return nil, NewNotDir(dirPath)
	}
	nodes := make([]FileInfo, 0, len(dir.children) + len(dir.files))
	for _, child := range dir.children {
		nodes = append(nodes, child)
	}
	for _, file := range dir.files {
		nodes = append(nodes, file)
	}
	return nodes, nil
}

func (fs *MemoryFS) ReadFile(filePath string, encoding string) (string, error) {
	if !isFilePath(filePath) {
		return "", NewNotFile(filePath)
	}
	filePath = fs.resolve(filePath)
	dir, file := fs.locate(filePath, false)
	if dir == nil {
		return "", NewFileOrDirNotExists(filePath)
	}
	if file == nil {
		return "", NewNotFile(filePath)
	}
	return file.content, nil
}

func (fs *MemoryFS) WriteFile(filePath string, fileText string) error {
	if !isFilePath(filePath) {
		return NewNotFile(filePath)
	}
	filePath = fs.resolve(filePath)
	dir, file := fs.locate(filePath, true)
	if dir == nil {
		dirPath := dirName(filePath)
		if dirPath == "" {
			dirPath = fs.current.FullPath()
		}
		return NewFileOrDirNotExists(dirPath)
	}
	now := time.Now()
	if file != nil {
		sizeDiff := int64(len(fileText)) - file.Size()
		file.content = fileText
		file.parent.size += sizeDiff
		file.modeTime = now
		file.parent.modeTime = now
	} else {
		fileName := baseName(filePath)
		if fileName == "" {
			return NewNotFile(filePath)
		}
		file = &MemoryFileNode{
			parent:   dir,
			name:     fileName,
			content:  fileText,
			modeTime: now,
		}
		dir.files[fileName] = file
		dir.size += file.Size()
		dir.modeTime = now
	}
	return nil
}

func (fs *MemoryFS) mkdir(dirPath string) (*MemoryDirNode, error) {
	dirPath = fs.resolve(dirPath)
	node := fs.root
	parts := strings.Split(dirPath, "/")
	now := time.Now()
	for _, part := range parts {
		if part != "" {
			dir, found := node.children[part]
			if !found {
				dir = &MemoryDirNode{
					parent:   node,
					name:     part,
					size:     0,
					modeTime: now,
				}
				if node.children == nil {
					node.children = make(map[string]*MemoryDirNode)
					node.children[dir.name] = dir
					node.modeTime = now
				}
			}
			node = dir
		}
	}
	return node, nil
}

func (fs *MemoryFS) Mkdir(dirPath string) error {
	_, err := fs.mkdir(dirPath)
	return err
}

func (fs *MemoryFS) copy(srcPath string, destPath string, remove bool) error {
	isSrcDir := strings.HasSuffix(srcPath, "/")
	isDestDir := strings.HasSuffix(destPath, "/")
	srcPath = fs.resolve(srcPath)
	destPath = fs.resolve(destPath)
	srcDir, srcFile := fs.locate(srcPath, false)
	now := time.Now()
	if srcDir == nil {
		return NewFileOrDirNotExists(srcPath)
	}
	if srcFile != nil {
		if isSrcDir {
			return NewFileOrDirNotExists(srcPath)
		}
		destDir, destFile := fs.locate(destPath, false)
		if destFile != nil && isDestDir {
			return NewNotDir(destPath)
		}
		if destDir == nil {
			destDir, _ = fs.mkdir(destPath)
		}
		if destFile == nil {
			destFile = &MemoryFileNode{
				parent:   destDir,
				name:     srcFile.name,
				content:  srcFile.content,
				modeTime: now,
			}
			destDir.files[destFile.name] = destFile
			destDir.size += destFile.Size()
		} else {
			sizeDiff := srcFile.Size() - destFile.Size()
			destFile.content = srcFile.content
			destFile.modeTime = now
			destDir.size += sizeDiff
		}
		destDir.modeTime = now
		if remove {
			srcFile.Delete()
		}
	} else {
		destDir, destFile := fs.locate(destPath, false)
		if destFile != nil {
			return NewNotDir(destPath)
		}
		if destDir == nil {
			destDir, _ = fs.mkdir(destPath)
		}
		if !remove {
			srcDir = srcDir.deepCopy(now)
		}
		destDir.merge(srcDir, true)
		if remove {
			srcDir.Delete()
		}
	}
	return nil
}

func (fs *MemoryFS) Move(srcPath string, destPath string, remove bool) error {
	return fs.copy(srcPath, destPath, true)
}

func (fs *MemoryFS) Copy(srcPath string, destPath string, remove bool) error {
	return fs.copy(srcPath, destPath, false)
}

func (fs *MemoryFS) FileExists(filePath string) (bool, error) {
	if !isFilePath(filePath) {
		return false, nil
	} else {
		filePath = fs.resolve(filePath)
		_, file := fs.locate(filePath, false)
		return file != nil, nil
	}
}

func (fs *MemoryFS) DirectoryExists(dirPath string) (bool, error) {
	dirPath = fs.resolve(dirPath)
	dir, file := fs.locate(dirPath, false)
	return dir != nil && file == nil, nil
}

func (fs *MemoryFS) Realpath(path string) (string, error) {
	return fs.resolve(path), nil
}

func (fs *MemoryFS) GetCurrentDirectory() (string, error) {
	return fs.current.FullPath(), nil
}

func isNotPattern(part string) bool {
	return !strings.ContainsAny(part, "?*[{\\") 
}

func (fs *MemoryFS) _glob(node *MemoryDirNode, parts []string) ([]string, error) {
	for i, part := range parts {
		if part != "" {
			if isNotPattern(part) {
				dir, found := node.children[part]
				if found {
					node = dir
					continue
				}
				file, found := node.files[part]
				if found && i == len(parts) - 1 {
					return []string {file.FullPath()}, nil
				} else {
					return nil, nil
				}
			} else {
				g, err := glob.Compile(part)
				if err != nil {
					return nil, err
				}
				var pathes []string
				for dirName, dir := range node.children {
					if g.Match(dirName) {
						res, err := fs._glob(dir, parts[i + 1:])
						if err != nil {
							return nil, err
						} else if res != nil {
							pathes = append(pathes, res...)
						}
					}
				}
				for fileName, file := range node.files {
					if g.Match(fileName) {
						pathes = append(pathes, file.FullPath())
					}
				}
				return pathes, nil
			}
		}
	}
	return nil, nil
}

func (fs *MemoryFS) glob(pattern string) ([]string, error) {
	for {
		if strings.HasSuffix(pattern, "/") {
			pattern = pattern[0 : len(pattern)-1]
		} else {
			break
		}
	}
	var node *MemoryDirNode
	if strings.HasPrefix(pattern, "/") {
		node = fs.root
	} else {
		node = fs.current
	}
	parts := strings.Split(pattern, "/")
	return fs._glob(node, parts)
}

func (fs *MemoryFS) Glob(patterns []string) ([]string, error) {
	var pathes []string
	for _, pattern := range patterns {
		res, err := fs.glob(pattern)
		if err != nil {
			return nil, err
		}
		pathes = append(pathes, res...)
	}
	return pathes, nil
}