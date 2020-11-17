package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

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

type AuthConfig struct {
	Type   string
	Static auth.StaticAuthConfig
	HTTP   auth.HTTPAuthConfig
}

// GetAuthenticator creates a new authenticator according to AuthConfig
func GetAuthenticator(conf AuthConfig) (auth.Authenticator, error) {
	switch conf.Type {
	case "static":
		return auth.NewStaticAuth(conf.Static), nil
	case "http":
		return auth.NewHTTPAuth(conf.HTTP), nil
	default:
		return nil, fmt.Errorf("Unknown auth type '%v'", conf.Type)
	}
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
			Type: "static",
			Static: auth.StaticAuthConfig{
				// Allow everything by default
				Allow: []string{"*"},
			},
			HTTP: auth.HTTPAuthConfig{
				URL:           "http://localhost:8080/publish",
				Timeout:       time.Second,
				Application:   "stream",
				PasswordParam: "auth",
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
