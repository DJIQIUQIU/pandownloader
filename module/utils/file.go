package utils

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func ReName(file string) string {
    count := 0
    fileDir := path.Dir(file)
    fileBase := path.Base(file)
    fileType := path.Ext(file)
    baseName := strings.Replace(fileBase, fileType, "", 1)
	for {
		if Exists(file) {
			count = count + 1
			file = fileDir + "/" + baseName + "(" + strconv.Itoa(count) + ")" + fileType
		} else {
			break
		}
	}
	fmt.Println("reName", file)
	return file
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func Move(source string, target string) bool {
	err := os.Rename(source, target)
	if err != nil {
        fmt.Printf("MvError: %v\n", err)
		return false
	}
	return true
}

/**
 * 创建目录
 */
func Mkdir(path string) bool {
	if !Exists(path) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

/**
 * 判断是否是目录
 */
func IsFile(f string) bool {
	fi, e := os.Stat(f)
	if e != nil {
		return false
	}
	return !fi.IsDir()
}
