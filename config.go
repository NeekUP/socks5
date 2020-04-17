package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

type config struct {
	Network string
	Address string
	Port    int
	Auth    AuthType
	User    []byte
	Pass    []byte
	MTU     int
}

type ymlconfig struct {
	Network string
	Address string
	Port    int
	Auth    string
	User    string
	Pass    string
	MTU     int
}

func tryParseConfig() (config, bool) {
	cfgFile, err := os.OpenFile(configFilename, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Unable to open config file %s: %s:", configFilename, err.Error())
		return config{}, false
	}

	var ymlcfg ymlconfig
	decoder := yaml.NewDecoder(cfgFile)
	err = decoder.Decode(&ymlcfg)
	if err != nil { // && err != io.EOF
		fmt.Printf("Fail to parse config: %v", err)
		return config{}, false
	}

	var auth AuthType
	switch strings.ToUpper(ymlcfg.Auth) {
	case "NO":
		auth = NO_AUTH
	case "PASS":
		auth = PASS_AUTH
		if ymlcfg.User == "" || ymlcfg.Pass == ""{
			fmt.Printf("User or password not defined")
			return config{}, false
		}
	default:
		fmt.Printf("Unknown auth type")
		return config{}, false
	}

	cfg := config{
		Network: ymlcfg.Network,
		Address: ymlcfg.Address,
		Port:    ymlcfg.Port,
		Auth:    auth,
		User:    []byte(ymlcfg.User),
		Pass:    []byte(ymlcfg.Pass),
		MTU:     ymlcfg.MTU,
	}

	return cfg, true
}
