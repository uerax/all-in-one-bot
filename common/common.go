package common

import (
	"archive/zip"
	"io"
	"os"
	"reflect"
)

func InSlice(slice, val any) bool {
	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return false
	}
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(val, v.Index(i).Interface()) {
            return true
        }
	}
	return false
}

func Zip(gifFile, outFile string) error {
	zipFile, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 打开源文件
	sourceFile, err := os.Open(gifFile)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 获取源文件的文件信息
	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// 创建zip文件中的文件
	zipEntry, err := zipWriter.Create(sourceFileInfo.Name())
	if err != nil {
		return err
	}

	// 将源文件内容拷贝到zip文件中
	_, err = io.Copy(zipEntry, sourceFile)
	if err != nil {
		return err
	}

	return nil
}