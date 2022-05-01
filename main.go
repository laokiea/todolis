package main

import (
	"log"

	"github.com/laokiea/todolist/cmd"
	"github.com/laokiea/todolist/list"
)

type pepper struct {
	Name     string
	HeatUnit int
	Peppers  int
}

func main() {
	//quit := make(chan<- os.Signal, 1)
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 初始化命令，也就是终端执行的进程
	if err := cmd.NewCommand().Execute(); err != nil {
		log.Fatal(err)
	}
	// 最后在进程退出前把todolist内容写到文件里保存
	list.GlobalLists.Flush()
}
