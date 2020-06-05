package main

import (
	"flag"
	"fmt"
	mp3_trans "github.com/dykily/mp3srt/mp3-trans"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	fmt.Println("hello world")
	//致命错误捕获
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("")
			log.Printf("错误:\n%v", err)

			time.Sleep(time.Second * 5)
		}
	}()

	appDir, err := filepath.Abs(filepath.Dir(os.Args[0])) //应用执行根目录
	if err != nil {
		panic(err)
	}

	//初始化
	if len(os.Args) < 2 {
		os.Args = append(os.Args , "")
	}

	var tranPath string

	//设置命令行参数
	flag.StringVar(&tranPath, "f", "", "enter a tranPath file waiting to be processed .")

	flag.Parse()

	if tranPath == "" && os.Args[1] != "" && os.Args[1] != "-f" {
		tranPath = os.Args[1]
	}
	fmt.Println(tranPath, appDir)
	app := mp3_trans.NewApp()




	// 调起应用
	app.RunMP3(tranPath)

	//延迟退出
	time.Sleep(time.Second * 1)
}
