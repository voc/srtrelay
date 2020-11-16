package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pelletier/go-toml"
	"github.com/voc/srtrelay/auth"
)

type Config struct {
	App  AppConfig
	Auth AuthConfig
}

type AppConfig struct {
	Address    string
	Port       uint
	Latency    uint
	Buffersize uint
}

type AuthType int

const (
	AuthTypeStatic AuthType = iota
	AuthTypeHttp
)

func (a *AuthType) UnmarshalTOML(src interface{}) error {
	log.Println("got", src)
	switch v := src.(type) {
	case string:
		if v == "static" {
			*a = AuthTypeStatic
		} else if v == "http" {
			*a = AuthTypeHttp
		} else {
			return fmt.Errorf("Unknown type '%s'", v)
		}
		return nil
	default:
	}
	return errors.New("Unknown type")
}

type AuthConfig struct {
	Type   AuthType
	Static auth.StaticAuthConfig
	Http   auth.HttpAuthConfig
}

func GetAuthenticator(conf AuthConfig) auth.Authenticator {
	if conf.Type == AuthTypeHttp {
		return auth.NewHttpAuth(conf.Http)
	}

	return auth.NewStaticAuth(conf.Static)
}

// Parse tries to find and parse config from paths in order
func Parse(paths []string) (*Config, error) {
	// set defaults
	config := Config{
		App: AppConfig{
			// TODO: see if we can make IPv6 or even dual socket work
			Address:    "127.0.0.1",
			Port:       1337,
			Latency:    300,
			Buffersize: 384000,
		},
		Auth: AuthConfig{
			Type: AuthTypeStatic,
			Static: auth.StaticAuthConfig{
				Allow: []string{"*"},
			},
		},
	}

	var data []byte
	var err error

	// try to read file from given paths
	for _, path := range paths {
		data, err = ioutil.ReadFile(path)
		if err == nil {
			log.Println("Read config from", path)
			break
		} else {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
	}

	// parse toml
	if data != nil {
		err = toml.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
	} else {
		log.Println("Config file not found, using defaults")
	}

	return &config, nil
}
