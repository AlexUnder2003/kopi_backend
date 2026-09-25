package main

import (
	"KopiBackend/internal/app"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	application := app.NewApp()

	go func() {
		application.Run()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	application.Stop()
}
