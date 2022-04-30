package main

import (
	"log"

	"github.com/laokiea/todolist/cmd"
)

func main() {
	//quit := make(chan<- os.Signal, 1)
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	cmd.NewCommand()
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
