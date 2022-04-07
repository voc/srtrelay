package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/pelletier/go-toml"
	"github.com/voc/srtrelay/auth"
)

type Config struct {
	App  AppConfig
	Auth AuthConfig
	API  APIConfig
}

type AppConfig struct {
	Address       string
	Addresses     []string
	PublicAddress string
	Latency       uint
	ListenTimeout uint
	Buffersize    uint
	SyncClients   bool
	LossMaxTTL    uint
}

type AuthConfig struct {
	Type   string
	Static auth.StaticAuthConfig
	HTTP   auth.HTTPAuthConfig
}

type APIConfig struct {
	Enabled bool
	Address string
	Port    uint
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

func getHostname() string {
	name, err := fqdn.FqdnHostname()
	if err != nil {
		log.Println("fqdn:", err)
		if err != fqdn.ErrFqdnNotFound {
			return name
		}

		name, err = os.Hostname()
		if err != nil {
			log.Println("hostname:", err)
		}
	}
	return name
}

// Parse tries to find and parse config from paths in order
func Parse(paths []string) (*Config, error) {
	// set defaults
	config := Config{
		App: AppConfig{
			Addresses:   []string{"localhost:1337"},
			Latency:     200,
			ListenTimeout: 3000,
			LossMaxTTL:  0,
			Buffersize:  384000,
			SyncClients: false,
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
		API: APIConfig{
			Enabled: true,
			Address: ":8080",
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

	// support old config files
	if config.App.Address != "" {
		log.Println("Note: config option address is deprecated, please use addresses")
		config.App.Addresses = []string{config.App.Address}
	}

	// guess public address if not set
	if config.App.PublicAddress == "" {
		split := strings.Split(config.App.Addresses[0], ":")
		if len(split) < 2 {
			log.Fatal("Invalid address: ", config.App.Addresses[0])
		}
		config.App.PublicAddress = fmt.Sprintf("%s:%s", getHostname(), split[len(split)-1])
		log.Println("Note: assuming public address", config.App.PublicAddress)
	}

	return &config, nil
}
