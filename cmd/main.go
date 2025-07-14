package main

import (
	"flag"

	"github.com/VortexSilence/X/config"
	"github.com/VortexSilence/X/handler"
)

func main() {
	configPath := flag.String("config", "config.json", "config file")
	flag.Parse()
	err := config.Load(*configPath)
	if err != nil {
	}
	go handler.Handle()
	for {
	}
}
