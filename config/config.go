package config

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/pelletier/go-toml/v2"
	"github.com/voc/srtrelay/auth"
)

const MetricsNamespace = "srtrelay"

type Config struct {
	App  AppConfig  `toml:"app"`
	Auth AuthConfig `toml:"auth"`
	API  APIConfig  `toml:"api"`
}

type AppConfig struct {
	// Deprecated, use Addresses
	DeprecatedAddress string `toml:"address"`

	// List of addresses to bind to
	Addresses []string `toml:"addresses"`

	// Address to use for API responses
	PublicAddress string `toml:"publicAddress"`

	// SRT LatencyMs in milliseconds
	LatencyMs uint `toml:"latency"`

	// total buffer size in bytes, determines maximum delay of a client
	Buffersize uint `toml:"buffersize"`

	// Whether to sync clients to GOP start
	SyncClients bool `toml:"syncClients"`

	// The value up to which the Reorder Tolerance may grow, 0 by default
	LossMaxTTL uint32 `toml:"lossMaxTTL"`

	// max size of packets in bytes, default is 1316
	PacketSize uint32 `toml:"packetSize"`
}

type AuthConfig struct {
	Type   string                `toml:"type"`
	Static auth.StaticAuthConfig `toml:"static"`
	HTTP   auth.HTTPAuthConfig   `toml:"http"`
}

type APIConfig struct {
	Enabled bool   `toml:"enabled"`
	Address string `toml:"address"`
	Port    uint   `toml:"port"`
}

// GetAuthenticator creates a new authenticator according to AuthConfig
func GetAuthenticator(conf AuthConfig) (auth.Authenticator, error) {
	switch conf.Type {
	case "static":
		return auth.NewStaticAuth(conf.Static), nil
	case "http":
		return auth.NewHTTPAuth(conf.HTTP), nil
	default:
		return nil, fmt.Errorf("unknown auth type '%v'", conf.Type)
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
			Addresses:  []string{"localhost:1337"},
			LatencyMs:  200,
			PacketSize: 1316, // max is 1456
		},
		Auth: AuthConfig{
			Type: "static",
			Static: auth.StaticAuthConfig{
				// Allow everything by default
				Allow: []string{"*"},
			},
			HTTP: auth.HTTPAuthConfig{
				URL:           "http://localhost:8080/publish",
				Timeout:       auth.Duration(time.Second),
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
		data, err = os.ReadFile(path)
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
	if config.App.DeprecatedAddress != "" {
		log.Println("Note: config option address is deprecated, please use addresses")
		config.App.Addresses = []string{config.App.DeprecatedAddress}
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

var (
	ipv4Loopback = netip.AddrFrom4([4]byte{127, 0, 0, 1})
	ipv6Loopback = netip.IPv6Loopback()
)

func ParseAddress(addr string) ([]netip.AddrPort, error) {
	// parse address
	addrStr, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		if _, ok := err.(*net.AddrError); ok {
			return nil, fmt.Errorf("invalid address: %q", addr)
		}
		return nil, err
	}

	// parse port
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %q", portStr)
	}

	if addrStr == "localhost" {
		return []netip.AddrPort{
			netip.AddrPortFrom(ipv4Loopback, uint16(portInt)),
			netip.AddrPortFrom(ipv6Loopback, uint16(portInt)),
		}, nil
	}

	// listen on both IPv4 and IPv6 if address is empty
	// (e.g. ":1337")
	if addrStr == "" {
		return []netip.AddrPort{
			netip.AddrPortFrom(netip.IPv6Unspecified(), uint16(portInt)),
		}, nil
	}

	addrPort, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, err
	}
	return []netip.AddrPort{addrPort}, nil
}
