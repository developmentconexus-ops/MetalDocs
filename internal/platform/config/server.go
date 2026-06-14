package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ServerConfig holds HTTP listen configuration.
type ServerConfig struct {
	Addr string
}

// LoadServerConfig reads APP_PORT from the environment.
// If unset, Addr defaults to ":8080".
// If set, the value must be a valid port in the range 1..65535.
func LoadServerConfig() (ServerConfig, error) {
	appPort := strings.TrimSpace(os.Getenv("APP_PORT"))
	if appPort == "" {
		return ServerConfig{Addr: ":8080"}, nil
	}
	port, err := strconv.Atoi(appPort)
	if err != nil || port < 1 || port > 65535 {
		return ServerConfig{}, fmt.Errorf("invalid APP_PORT value")
	}
	return ServerConfig{Addr: ":" + strconv.Itoa(port)}, nil
}
