package util

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

//FileWrite used to create a file and write to it, if file exists it will be overridden
func FileWrite(filePath, content string) (n int, err error) {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Unable tto create file for writing")
	}
	defer f.Close()
	return f.WriteString(content)
}

//FileRead reads contents of a file at a particular path
func FileRead(filePath string) ([]byte, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(b), nil
}

//CreateDir creats a directory
func CreateDir(dirPath string, perm os.FileMode) (err error) {
	err = os.MkdirAll(dirPath, perm)

	if err != nil {
		fmt.Println("failed to create directory " + dirPath + " error " + err.Error())
	}

	return err

}
