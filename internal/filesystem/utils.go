package filesystem

import "strings"

func cleanHeadSlash(path string) string {
	for {
		if strings.HasPrefix(path, "/") {
			path = path[1:]
		} else {
			break
		}
	}
	return path
}

func dirName(filePath string) string {
	i := strings.LastIndex(filePath, "/")
	if i == -1 {
		return ""
	} else {
		dir := filePath[0:i]
		if dir == "" {
			return "/"
		} else {
			return dir
		}
	}
}

func baseName(filePath string) string {
	i := strings.LastIndex(filePath, "/")
	if i == -1 {
		return filePath
	} else {
		return filePath[i:]
	}
}