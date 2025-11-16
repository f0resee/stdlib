package extension

import "github.com/f0resee/stdlib/apollo/env/file"

var fileHandler file.FileHandler

func SetFileHandler(inFile file.FileHandler) {
	fileHandler = inFile
}

func GetFileHandler() file.FileHandler {
	return fileHandler
}
