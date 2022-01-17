package config

import (
	"encoding/json"
	"os"
)

type MQTTConf struct {
	Address string
	Port    int
	User    string
	Pass    string
}

type AppConf struct {
	Devices    []string
	SampleSize int
	DeviceTTL  int
	MQTT       *MQTTConf
	Name       string
}

var globalConf *AppConf

func Get() *AppConf {
	if globalConf == nil {
		globalConf = readConf()
	}

	return globalConf
}

func readConf() *AppConf {
	dat, err := os.ReadFile("roomystan-conf.json")
	if err != nil {
		panic("failed to read ./roomystan-conf.json")
	}

	cfg := &AppConf{}

	err = json.Unmarshal(dat, cfg)
	if err != nil {
		panic("failed to unmarshal config")
	}

	sensibleDefaults(cfg)

	return cfg
}

func sensibleDefaults(cfg *AppConf) {
	if cfg.SampleSize == 0 {
		cfg.SampleSize = 3
	}

	if cfg.DeviceTTL == 0 {
		cfg.DeviceTTL = 30
	}

	if cfg.Name == "" {
		cfg.Name = "local"
	}

	if cfg.MQTT == nil {
		cfg.MQTT = &MQTTConf{
			Address: "localhost",
			Port:    1883,
			User:    "",
			Pass:    "",
		}
	}
}
