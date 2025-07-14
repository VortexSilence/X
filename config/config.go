package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Config struct {
	PMode      string         `json:"p_mode"`
	Inbounds   []Inbound      `json:"inbounds"`
	Outbounds  []Outbound     `json:"outbounds"`
	Mode       string         `json:"mode"`
	Auth       string         `json:"auth"`
	InMode     string         `json:"in_mode"`
	Port       int            `json:"port"`
	ToPort     int            `json:"to_port"`
	TunMode    string         `json:"tun_mode"`
	Type       string         `json:"type"`
	Transport  string         `json:"transport"`
	TCPConfig  map[string]any `json:"tcpConfig"`
	WSConfig   map[string]any `json:"wsConfig"`
	GRPCConfig map[string]any `json:"grpcConfig"`
	Pipe       string         `json:"pipe"`
	PipeConfig map[string]any `json:"pipeConfig"`
	Ports      []string       `json:"ports"`
}

type Inbound struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

type Outbound struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

var (
	cfg  Config
	once sync.Once
	err  error
)

// Load only once
func Load(path string) error {
	once.Do(func() {
		var configFile *os.File
		configFile, err = os.Open(path)
		if err != nil {
			err = fmt.Errorf("failed to open config file: %w", err)
			return
		}
		defer configFile.Close()

		jsonParser := json.NewDecoder(configFile)
		if e := jsonParser.Decode(&cfg); e != nil {
			err = fmt.Errorf("failed to decode JSON: %w", e)
		}
	})
	return err
}

// Get returns the loaded config
func Get() Config {
	return cfg
}
