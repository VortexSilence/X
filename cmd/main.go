package main

import (
	"flag"

	"core/config"
	"core/handler"
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
