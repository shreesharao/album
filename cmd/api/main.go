package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log/level"
	"github.com/shreesharao/album/internal/handler"
	"github.com/shreesharao/album/internal/resources"
)

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		done <- true
	}()

	level.Info(resources.Log).Log("msg", "starting the application.")
	err := handler.StartListening(resources.Log, resources.KafkaProducer)
	if err != nil {
		level.Error(resources.Log).Log("msg", "error on starting the application.", "err", err)
		os.Exit(1)
	}
	<-done
}
