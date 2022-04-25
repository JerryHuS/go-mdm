package qm

import (
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

//
// 获取进程路径
//
func GetExeFilePath() string {
	path, err := filepath.Abs(os.Args[0])
	if err != nil {
		LOG_ERROR(err.Error())
	}
	return path
}

//
// 获取进程名，带后缀
//
func GetExeFileName() string {
	path, err := filepath.Abs(os.Args[0])

	if err != nil {
		return ""
	}

	return filepath.Base(path)
}

//
// 获取进程名，不带后缀
//
func GetExeFileBaseName() string {
	name := GetExeFileName()
	return strings.TrimSuffix(name, filepath.Ext(name))
}

//
// 路径最后添加反斜杠
//
func PathAddBackslash(path string) string {
	i := len(path) - 1;
	if !os.IsPathSeparator(path[i]) {
		path += string(os.PathSeparator)
	}

	return path
}

//
// 路径最后去除反斜杠
//
func PathRemoveBackslash(path string) string {
	i := len(path) - 1

	if i > 0 && os.IsPathSeparator(path[i]) {
		path = path[:i]
	}

	return path
}

func PathFileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

//
// 获取进程所在目录: 末尾带反斜杠
//
func GetMainDiectory() (string) {
	path, err := filepath.Abs(os.Args[0])

	if err != nil {
		return ""
	}

	full_path := filepath.Dir(path)

	return PathAddBackslash(full_path)
}

//
// 获取进程所在目录的文件路径：filename表示文件名
//
func GetMainPath(filename string) string {
	return GetMainDiectory() + filename
}

//
// 创建目录
//
func CreateDirectory(path string) bool {
	return os.Mkdir(path, os.ModePerm) == nil
}

// @hwy
// 复制文件,目标文件路径不存在会创建
// srcpath string: 源文件
// destpath string: 目标文件
func CopyFile(srcpath string, destpath string) bool {
	src, err := os.Open(srcpath)
	if err != nil {
		LOG_ERROR_F("Failed to open srcfile, filePath:%s, err:%v.", srcpath, err)
		return false
	}
	defer src.Close()

	// 创建目标文件父目录
	destDir := filepath.Dir(destpath)
	if !PathFileExists(destDir) {
		os.MkdirAll(destDir, os.ModePerm)
	}
	dst, err := os.OpenFile(destpath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		LOG_ERROR_F("Failed to open destfile, filePath:%s, err:%v.", destpath, err)
		return false
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		LOG_ERROR_F("Failed to copy file, srcfile:%s, destfile:%s, err:%v.", srcpath, destpath, err)
		return false
	}
	return true
}

// @hwy
// 获取文件名，包括后缀。如：/a/s/c.txt->c.txt; /a/s/c.txt/->空串; abs->abs;
//
func GetPathFileName(path string) string {
	path = strings.Replace(path, "\\", "/", -1)
	findex := strings.LastIndex(path, "/")

	return path[findex+1:]
}

// @hwy
// 删除目录中的文件名包含指定字符串的文件
// 返回值:
//   bool:删除失败返回false
func RemoveFile(path string, match string) bool {
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if strings.Index(path, match) != -1 {
			return os.Remove(path)
		}
		return nil
	})

	if err != nil {
		return false
	}
	return true
}

// @hwy
// 清空文件夹中所有的内容，不删除该文件夹
func ClearDir(path string) bool {
	if !PathFileExists(path) {
		LOG_INFO_F("Folder does not exist, path:%s.", path)
		return true
	}
	lists, err := ioutil.ReadDir(path)
	if err != nil {
		LOG_ERROR_F("Failed to readDir, dir:%s.", path)
		return false
	}

	for _, fi := range lists {
		curpath := path + string(os.PathSeparator) + fi.Name()
		err := os.RemoveAll(curpath)
		if err != nil {
			LOG_INFO_F("Failed to execute removeAll, path:%s.", curpath)
			continue
		}
	}
	return true
}

// 拷贝文件夹
//
func CopyDir(srcDir string, destDir string) bool {
	srcDir, _ = filepath.Abs(srcDir)
	destDir, _ = filepath.Abs(destDir)
	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {

		} else {
			newDest := strings.Replace(path, srcDir, destDir, -1)
			LOG_INFO_F("CopyFile, srcFile:%s, destFile:%s.", path, destDir+"/"+f.Name())
			CopyFile(path, newDest)
		}
		return nil
	})
	if err != nil {
		LOG_ERROR_F("Failed to copyDir, err:%v.", err)
		return false
	}
	return true
}

// 移动源文件夹所有文件到目标文件夹中
func MoveDir(srcDir string, destDir string) bool {
	srcDir, _ = filepath.Abs(srcDir)
	destDir, _ = filepath.Abs(destDir)
	err := filepath.Walk(srcDir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {

		} else {
			newDest := strings.Replace(path, srcDir, destDir, -1)
			newDestDir := filepath.Dir(newDest)
			if !PathFileExists(newDestDir) {
				err := os.MkdirAll(newDestDir, os.ModePerm)
				if err != nil {
					LOG_ERROR_F("Failed to os.MkdirAll(%s), err:%v.", newDestDir, err)
					return errors.New("Failed to os.MkdirAll.")
				}
			}
			err := os.Rename(path, newDest)
			if err != nil {
				LOG_ERROR_F("Failed to move file, src:%s, dest:%s, err:%v.", path, newDest, err)
				return errors.New("Failed to move file.")
			}
			LOG_INFO_F("MoveFile, srcFile:%s, destFile:%s.", path, newDest)
		}
		return nil
	})
	if err != nil {
		LOG_ERROR_F("Failed to moveDir, err:%v.", err)
		return false
	}
	return true
}

/**
* description: 判断路径是否存在和是否为文件
* -----------
* para: path: 路径
* -----------
* return:
* bool: isExist: 是否存在
* bool: isFile: 是否为文件
 */
func IsPathFileExist(path string) (isExist bool, isFile bool) {
	isFile = false
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, isFile
	}
	isExist = true
	isFile = !fileInfo.IsDir()
	return
}

func MoveFile(srcFile string, dstFile string) error {
	//
	isFileExist, isFile := IsPathFileExist(srcFile)
	if !isFileExist || !isFile {
		return errors.New("srcFile is not exist or not a file.")
	}
	//
	newDestDir := filepath.Dir(dstFile)
	if !PathFileExists(newDestDir) {
		err := os.MkdirAll(newDestDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	err := os.Rename(srcFile, dstFile)
	if err != nil {
		return err
	}
	return nil
}

func ReadFileAsString(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}