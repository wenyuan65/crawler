package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// 将lines按行写入文件中， append表示是否追加写模式
func WriteToFile(filePath string, lines []string, append bool) {
	flag := os.O_RDWR | os.O_CREATE
	if append {
		flag = flag | os.O_APPEND
	}

	file, err := os.OpenFile(filePath, flag, 0666)
	if err != nil {
		fmt.Printf("打开文件错误=%s\r\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		writer.WriteString(line)
		writer.WriteString("\r\n")
	}
	writer.Flush()
}

func ReadLines(filePath string) []string {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("打开文件错误=%v \n", err)
		return []string{}
	}
	defer f.Close()

	lines := []string{}
	r := bufio.NewReader(f)
	for {
		line, _, err2 := r.ReadLine()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			fmt.Printf("读取文件错误=%v \r\n", err)
			return []string{}
		}

		lines = append(lines, string(line))
	}

	return lines
}

// 判断文件或文件夹否存在
func IsFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		panic(err)
	}
}
