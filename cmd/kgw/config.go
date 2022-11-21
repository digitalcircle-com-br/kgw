package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ConfigRoute struct {
	Path      string `yaml:"path"`
	Target    string `yaml:"target"`
	StripPath bool   `yaml:"strip-path"`
}

type ConfigAcme struct {
	Dir       string   `yaml:"dir"`
	Email     string   `yaml:"email"`
	Enabled   bool     `yaml:"enabled"`
	Whitelist []string `yaml:"whitelist"`
}

type Config struct {
	Addr     string        `yaml:"addr"`
	Secure   bool          `yaml:"secure"`
	LogLevel string        `yaml:"log-level"`
	Routes   []ConfigRoute `yaml:"routes"`
	Acme     *ConfigAcme   `yaml:"acme"`
}

var configs = []string{
	"./config.yaml",
	"/kgw/config.yaml",
	"/kgw/etc/config.yaml",
	"./etc/config.yaml",
	"/config.yaml",
}
var configName = ""

func detectConfigOnce() error {
	var err error
	if configName == "" {
		for _, f := range configs {
			_, err = os.Stat(f)
			if err == nil {
				logrus.Infof("Using config: %s", f)
				configName = f
				break
			}

		}
		if err != nil {
			return err
		}
	}

	bs, err := os.ReadFile(configName)
	if err != nil {
		return err
	}
	if string(bs) != string(lastCfg) {
		lastCfg = bs
		logrus.Infof("new cfg detected: %s", string(bs))
		err = buildMux()
		if err != nil {
			return err
		}
	}
	return nil
}

func detectConfig() {
	for {
		err := detectConfigOnce()
		if err != nil {
			logrus.Debugf("error reading config: %s", err.Error())
		}
		time.Sleep(time.Second * 5)
	}
}
