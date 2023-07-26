package main

import (
	"dusnet/logger"

	"github.com/jasonlvhit/gocron"
)

func main() {
	gocron.Every(1).Days().At("00:00:00").Do(func() {
		logger.Info("test corn....")
	})
	select {}
}
