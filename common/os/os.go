package os

import (
	"fmt"
	"io"
	"os"
)

// Exit ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Exit 发生了极其糟糕的错误，只能停止程序。
func Exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API这里定义的全是项目级全局函数

// FileExists ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FileExists 给定文件路径，判断文件存不存在。
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// EnsureDir ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// EnsureDir 给定一个文件夹的路径，如果该文件夹不存在，就新建一个，这个可以创建多级文件夹，
// 例如：go/src/meta--。
func EnsureDir(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return fmt.Errorf("os.EnsureDir: failed create directory %q for %q", dir, err)
	}
	return nil
}

// WriteFile ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// WriteFile 给定一个文件路径和待写入的内容，如果指定文件不存在，那么就新建一个文件，然后再将
// 内容写进去。
func WriteFile(filePath string, content []byte, mode os.FileMode) error {
	return os.WriteFile(filePath, content, mode)
}

// MustWriteFile ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// MustWriteFile 方法调用 WriteFile 方法，将给定的内容写入到指定文件里，如果发生了错误，就退出程序。
func MustWriteFile(filePath string, content []byte, mode os.FileMode) {
	if err := WriteFile(filePath, content, mode); err != nil {
		Exit(fmt.Errorf("os.MustWriteFile: failed to write content into %q for %q", filePath, err))
	}
}

// ReadFile ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ReadFile 给定文件的地址，从中读取数据出来。
func ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// CopyFile ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CopyFile 将给定的文件复制到指定的地方去。
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	defer func() {
		_ = srcFile.Close()
	}()
	if err != nil {
		return fmt.Errorf("os.CopyFile: failed to open source file %q for %q", src, err)
	}
	info, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("os.CopyFile: cannot copy directory %q", src)
	}
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_APPEND|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("os.CopyFile: failed to open destination file %q for %q", dst, err)
	}
	defer func() {
		_ = dstFile.Close()
	}()
	_, err = io.Copy(dstFile, srcFile)
	return err
}
