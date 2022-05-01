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
	if err := cmd.NewCommand().Execute(); err != nil {
		log.Fatal(err)
	}
	list.GlobalLists.Flush()
}
