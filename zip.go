package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// walkDirFunc 递归遍历目录，并将文件添加到zip.Writer中
func walkDirFunc(baseDir, dir string, zw *zip.Writer) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		fullPath := filepath.Join(dir, file.Name())

		if file.IsDir() {
			if err := walkDirFunc(baseDir, fullPath, zw); err != nil {
				return err
			}
		} else {
			if err := addFileToZip(baseDir, fullPath, zw); err != nil {
				return err
			}
		}
	}
	return nil
}

// addFileToZip 添加单个文件到zip文件
func addFileToZip(baseDir, file string, zw *zip.Writer) error {
	fileToZip, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// 获取文件相对于baseDir的相对路径作为zip内的路径
	relPath, err := filepath.Rel(baseDir, file)
	if err != nil {
		return err
	}

	// 创建zip文件内的文件信息头
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = relPath
	header.Method = zip.Deflate // 使用Deflate压缩算法

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

// zipFolder 压缩文件夹
func zipFolder(folderPath, zipFilePath string) error {
	newZipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zw := zip.NewWriter(newZipFile)
	defer zw.Close()

	return walkDirFunc(folderPath, folderPath, zw)
}

func zipFile(filePath, zipFilePath string) error {
	if err := zipFolder(filePath, zipFilePath); err != nil {
		fmt.Println("压缩zip错误:", err)
		return err
	} else {
		LogInfo("成功压缩:", zipFilePath)
		return nil
	}
}
