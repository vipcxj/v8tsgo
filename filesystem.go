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
	cbRealpath     func(path string) (string, error)

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

func NewV8FileSystem(fs filesystem.FileSystem, utils *V8Utils) (*V8FileSystemHost, error) {
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

	ot := v8.NewObjectTemplate(iso)
	v8.NewFunctionTemplate(iso, func(info *v8.FunctionCallbackInfo) *v8.Value {})
	return nil
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
