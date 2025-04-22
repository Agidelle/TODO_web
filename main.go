package main

import (
	"github.com/agidelle/todo_web/cmd"

	"os"
	"os/signal"
	"syscall"
)

func main() {
	app := cmd.Initialize()
	server := app.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	app.Stop(server)
}
