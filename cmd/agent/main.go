package main

import (
	"os"
	"os/signal"
	"syscall"

	"KopiBackend/internal/app"
)

func main() {
	application := app.NewAgent()

	go func() {
		application.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Stop()
}
