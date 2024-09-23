package v8tsgo

import (
	"fmt"
	"io/fs"

	"github.com/vipcxj/v8tsgo/internal/filesystem"
	v8 "rogchap.com/v8go"
)

type V8FileSystemHost struct {
	ctx   *v8.Context
	utils *V8Utils

	// Gets if this file system is case sensitive.
	//   isCaseSensitive(): boolean
	fnIsCaseSensitive *v8.FunctionTemplate

	// Asynchronously deletes the specified file or directory.
	//   delete(path: string): Promise<void>
	fnDelete *v8.FunctionTemplate
	// Synchronously deletes the specified file or directory.
	//   deleteSync(path: string): void
	fnDeleteSync *v8.FunctionTemplate

	// Reads all the child directories and files.
	// Implementers should have this return the full file path.
	//   readDirSync(dirPath: string): RuntimeDirEntry[]
	fnReadDirSync *v8.FunctionTemplate

	// Asynchronously reads a file at the specified path.
	//	 readFile(filePath: string, encoding?: string): Promise<string>
	fnReadFile *v8.FunctionTemplate
	// Synchronously reads a file at the specified path.
	//	 readFileSync(filePath: string, encoding?: string): string
	fnReadFileSync *v8.FunctionTemplate

	// Asynchronously writes a file to the file system.
	//   writeFile(filePath: string, fileText: string): Promise<void>
	fnWriteFile *v8.FunctionTemplate
	// Synchronously writes a file to the file system.
	//   writeFileSync(filePath: string, fileText: string): void
	fnWriteFileSync *v8.FunctionTemplate

	// Asynchronously creates a directory at the specified path.
	//   mkdir(dirPath: string): Promise<void>
	fnMkdir *v8.FunctionTemplate
	// Synchronously creates a directory at the specified path.
	//   mkdirSync(dirPath: string): void
	fnMkdirSync *v8.FunctionTemplate

	// Asynchronously moves a file or directory.
	//   move(srcPath: string, destPath: string): Promise<void>
	fnMove *v8.FunctionTemplate
	// Synchronously moves a file or directory.
	//   moveSync(srcPath: string, destPath: string): void
	fnMoveSync *v8.FunctionTemplate

	// Asynchronously copies a file or directory.
	//   copy(srcPath: string, destPath: string): Promise<void>
	fnCopy *v8.FunctionTemplate
	// Synchronously copies a file or directory.
	//   copySync(srcPath: string, destPath: string): void
	fnCopySync *v8.FunctionTemplate

	// Asynchronously checks if a file exists.
	// Implementers should throw an `errors.FileNotFoundError` when it does not exist.
	//   fileExists(filePath: string): Promise<boolean>
	fnFileExists *v8.FunctionTemplate
	// Synchronously checks if a file exists.
	// Implementers should throw an `errors.FileNotFoundError` when it does not exist.
	//   fileExistsSync(filePath: string): boolean
	fnFileExistsSync *v8.FunctionTemplate

	// Asynchronously checks if a directory exists.
	//   directoryExists(dirPath: string): Promise<boolean>
	fnDirectoryExists *v8.FunctionTemplate
	// Synchronously checks if a directory exists.
	//   directoryExistsSync(dirPath: string): boolean
	fnDirectoryExistsSync *v8.FunctionTemplate

	// See https://nodejs.org/api/fs.html#fs_fs_realpathsync_path_options
	//   realpathSync(path: string): string
	fnRealpathSync *v8.FunctionTemplate

	// Gets the current directory of the environment.
	//   getCurrentDirectory(): string
	fnGetCurrentDirectory *v8.FunctionTemplate

	// Uses pattern matching to find files or directories.
	//   glob(patterns: ReadonlyArray<string>): Promise<string[]>
	fnGlob *v8.FunctionTemplate
	// Synchronously uses pattern matching to find files or directories.
	//   globSync(patterns: ReadonlyArray<string>): string[]
	fnGlobSync *v8.FunctionTemplate
}

func extractArg(info *v8.FunctionCallbackInfo, index int) (*v8.Value, error) {
	if index >= len(info.Args()) {
		return nil, fmt.Errorf("not enough arguments, expect at least %d args, but got %d args", index+1, len(info.Args()))
	}
	return info.Args()[index], nil
}

func extractStringArg(info *v8.FunctionCallbackInfo, index int) (string, error) {
	value, err := extractArg(info, index)
	if err != nil {
		return "", err
	}
	if value.IsString() || value.IsStringObject() {
		return value.String(), nil
	} else {
		return "", fmt.Errorf("the arg %d is not a string or string object", index)
	}
}

func extractOptArg(info *v8.FunctionCallbackInfo, index int) *v8.Value {
	if index >= len(info.Args()) {
		return nil
	}
	return info.Args()[index]
}

func extractOptStringArg(info *v8.FunctionCallbackInfo, index int) (string, bool, error) {
	value := extractOptArg(info, index)
	if value == nil {
		return "", false, nil
	}
	if value.IsString() || value.IsStringObject() {
		return value.String(), true, nil
	} else {
		return "", false, fmt.Errorf("the arg %d is not a string or string object", index)
	}
}

func extractStringsArg(info *v8.FunctionCallbackInfo, index int) ([]string, error) {
	value, err := extractArg(info, index)
	if err != nil {
		return nil, err
	}
	if value.IsString() || value.IsStringObject() {
		return []string{value.String()}, nil
	} else {
		var out []string
		err = ParseValue(info.Context(), value, &out)
		if err != nil {
			return nil, fmt.Errorf("the arg %d is not a string or string array", index)
		}
		return out, nil
	}
}

func mustMakeResolver(ctx *v8.Context) *v8.PromiseResolver {
	resolver, err := v8.NewPromiseResolver(ctx)
	if err != nil {
		panic(err)
	}
	return resolver
}

func mustWrapError(utils *V8Utils, err error) *v8.Value {
	ex, err := utils.WrapError(err)
	if err != nil {
		panic(err)
	}
	return ex
}

func mustNewValue(iso *v8.Isolate, v any) *v8.Value {
	res, err := v8.NewValue(iso, v)
	if err != nil {
		panic(err)
	}
	return res
}

func mustMakeValue(ctx *v8.Context, v any) *v8.Value {
	value, err := MakeValue(ctx, v)
	if err != nil {
		panic(err)
	}
	return value
}

func toRuntimeDirEntry(info fs.FileInfo) map[string]any {
	return map[string]any{
		"name": info.Name(),
		"isFile": info.Mode().IsRegular(),
		"isDirectory": info.IsDir(),
		"isSymlink": info.Mode() & fs.ModeSymlink != 0,
	}
}

func NewV8FileSystem(fs filesystem.FileSystem, utils *V8Utils) *V8FileSystemHost {
	ctx := utils.ctx
	fsh := &V8FileSystemHost{
		ctx:   ctx,
		utils: utils,
	}
	iso := ctx.Isolate()
	fsh.fnIsCaseSensitive = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		res, err := v8.NewValue(iso, fs.IsCaseSensitive())
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		} else {
			return res
		}
	})
	fsh.fnCopy = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		srcPath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		destPath, err := extractStringArg(info, 1)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			err := fs.Copy(srcPath, destPath)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(v8.Undefined(iso))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnCopySync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		srcPath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		destPath, err := extractStringArg(info, 1)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		err = fs.Copy(srcPath, destPath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		} else {
			return v8.Undefined(iso)
		}
	})
	fsh.fnDelete = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		path, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			err := fs.Delete(path)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(v8.Undefined(iso))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnDeleteSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		path, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		err = fs.Delete(path)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		} else {
			return v8.Undefined(iso)
		}
	})
	fsh.fnDirectoryExists = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		dirPath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			res, err := fs.DirectoryExists(dirPath)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(mustNewValue(iso, res))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnDirectoryExistsSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		dirPath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		res, err := fs.DirectoryExists(dirPath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return mustNewValue(iso, res)
	})
	fsh.fnFileExists = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			res, err := fs.FileExists(filePath)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(mustNewValue(iso, res))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnFileExistsSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		res, err := fs.FileExists(filePath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return mustNewValue(iso, res)
	})
	fsh.fnGetCurrentDirectory = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		path, err := fs.GetCurrentDirectory()
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		} else {
			return mustNewValue(iso, path)
		}
	})
	fsh.fnGlob = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		patterns, err := extractStringsArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			res, err := fs.Glob(patterns)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(mustMakeValue(ctx, res))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnGlobSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		patterns, err := extractStringsArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		res, err := fs.Glob(patterns)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return mustMakeValue(ctx, res)
	})
	fsh.fnMkdir = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		dirPath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			err := fs.Mkdir(dirPath)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(v8.Undefined(iso))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnMkdirSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		dirPath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		err = fs.Mkdir(dirPath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return v8.Undefined(iso)
	})
	fsh.fnMove = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		srcPath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		destPath, err := extractStringArg(info, 1)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func() {
			err := fs.Move(srcPath, destPath)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(v8.Undefined(iso))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnMoveSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		srcPath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		destPath, err := extractStringArg(info, 1)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		err = fs.Move(srcPath, destPath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		} else {
			return v8.Undefined(iso)
		}
	})
	fsh.fnReadDirSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		dirPath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		infoes, err := fs.ReadDir(dirPath)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		result := make([]map[string]any, 0, len(infoes))
		for _, info := range infoes {
			result = append(result, toRuntimeDirEntry(info))
		}
		resValue, err := MakeValue(info.Context(), result)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return resValue
	})
	fsh.fnReadFile = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		encoding, ok, err := extractOptStringArg(info, 1)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		if !ok {
			encoding = "utf-8"
		}
		go func ()  {
			content, err := fs.ReadFile(filePath, encoding)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(mustNewValue(iso, content))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnReadFileSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		encoding, ok, err := extractOptStringArg(info, 1)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		if !ok {
			encoding = "utf-8"
		}
		content, err := fs.ReadFile(filePath, encoding)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return mustNewValue(iso, content)
	})
	fsh.fnRealpathSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		path, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		res, err := fs.Realpath(path)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return mustNewValue(iso, res)
	})
	fsh.fnWriteFile = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		resolver := mustMakeResolver(ctx)
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		fileText, err := extractStringArg(info, 1)
		if err != nil {
			resolver.Reject(mustWrapError(utils, err))
			return resolver.GetPromise().Value
		}
		go func ()  {
			err := fs.WriteFile(filePath, fileText)
			if err != nil {
				resolver.Reject(mustWrapError(utils, err))
			} else {
				resolver.Resolve(v8.Undefined(iso))
			}
		}()
		return resolver.GetPromise().Value
	})
	fsh.fnWriteFileSync = v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {
		filePath, err := extractStringArg(info, 0)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		fileText, err := extractStringArg(info, 1)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		err = fs.WriteFile(filePath, fileText)
		if err != nil {
			return iso.ThrowException(mustWrapError(utils, err))
		}
		return v8.Undefined(iso)
	})
	return fsh
}

func setMethod(target *v8.ObjectTemplate, name string, method *v8.FunctionTemplate) error {
	err := target.Set(name, method, v8.ReadOnly)
	if err != nil {
		return fmt.Errorf("failed to set method \"%s\" on FileSystemHost, %w", name, err)
	} else {
		return nil
	}
}

func (fs *V8FileSystemHost) CreateObjectTemplate() (*v8.ObjectTemplate, error) {
	t := v8.NewObjectTemplate(fs.ctx.Isolate())
	err := setMethod(t, "copy", fs.fnCopy)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "copySync", fs.fnCopySync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "delete", fs.fnDelete)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "deleteSync", fs.fnDeleteSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "directoryExists", fs.fnDirectoryExists)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "directoryExistsSync", fs.fnDirectoryExistsSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "fileExists", fs.fnFileExists)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "fileExistsSync", fs.fnFileExistsSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "getCurrentDirectory", fs.fnGetCurrentDirectory)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "glob", fs.fnGlob)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "globSync", fs.fnGlobSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "isCaseSensitive", fs.fnIsCaseSensitive)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "mkdir", fs.fnMkdir)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "mkdirSync", fs.fnMkdirSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "move", fs.fnMove)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "moveSync", fs.fnMoveSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "readDirSync", fs.fnReadDirSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "readFile", fs.fnReadFile)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "readFileSync", fs.fnReadFileSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "realpathSync", fs.fnRealpathSync)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "writeFile", fs.fnWriteFile)
	if err != nil {
		return nil, err
	}
	err = setMethod(t, "writeFileSync", fs.fnWriteFileSync)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (fs *V8FileSystemHost) CreateInstance() (*v8.Value, error) {
	t, err := fs.CreateObjectTemplate()
	if err != nil {
		return nil, fmt.Errorf("unable to create object template of FileSystemHost, %w", err)
	}
	v, err := t.NewInstance(fs.ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to instance of FileSystemHost, %w", err)
	}
	return v.Value, nil
}

func (f *V8FileSystemHost) isCaseSensitive() bool {
	return true
}

func (f *V8FileSystemHost) delete(path string) error {
	return nil
}

func (f *V8FileSystemHost) readDir(dirPath string) ([]fs.File, error) {
	return nil, nil
}
