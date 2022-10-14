package config

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Configuration struct {
	config map[string]string
}

var configuration *Configuration

func init() {
	configuration = &Configuration{}
	configuration.config = make(map[string]string)

	var files []string = []string{}
	err := filepath.Walk("./conf/", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	for i := 0; i < len(files); i++ {
		Load(files[i])
	}
}

func GetValue(key string) string {
	return configuration.config[key]
}

func GetIntValue(key string) int {
	value := configuration.config[key]
	intValue, _ := strconv.Atoi(value)

	return intValue
}

func ContainsKey(key string) bool {
	_, contains := configuration.config[key]
	return contains
}

func Load(path string) {
	fmt.Println("正在加载配置文件：" + path)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)
	for {
		line, _, err2 := r.ReadLine()
		if err2 != nil {
			if err2 == io.EOF {
				break
			}
			panic(err2)
		}
		s := strings.TrimSpace(string(line))
		// 空白行
		if len(s) == 0 {
			continue
		}
		// 注释
		if s[0:1] == "#" {
			continue
		}

		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}

		key := strings.TrimSpace(s[0:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}

		oldValue := configuration.config[key]
		if len(oldValue) > 0 {
			fmt.Printf("配置：%s = %s 被覆盖\r\n", key, oldValue)
		}

		configuration.config[key] = value
	}

	return
}
