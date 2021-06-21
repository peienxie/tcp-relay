package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alyu/configparser"
)

type Config struct {
	MiddleServerPort    int
	SecuredMiddleServer bool
	RelayTargetType           RelayTargetType
	RelayTargetAddress  string
	SecuredRelayTarget  bool
}

type RelayTargetType int

var (
	REDIRECT  RelayTargetType = 1
	LISTENING RelayTargetType = 2
)

func (t RelayTargetType) String() string {
	if t == REDIRECT {
		return "Redirect"
	} else if t == LISTENING {
		return "Listening"
	}
	return "unknown type"
}

func (c Config) String() string {
	return fmt.Sprintf("MiddleServerPort: %d\nSecuredMiddleServer: %v\nRelayTargetType: %s\nRelayTargetAddress: %s\nSecuredRelayTarget: %v\n\n",
		c.MiddleServerPort,
		c.SecuredMiddleServer,
		c.RelayTargetType,
		c.RelayTargetAddress,
		c.SecuredRelayTarget,
	)
}

func (c *Config) setupRelayTargetConfig(relayTypeStr string, configINI *configparser.Configuration) error {
	switch strings.ToLower(relayTypeStr) {
	case "redirect":
		target, err := configINI.Section("Redirect")
		if err != nil {
			return err
		}
		ip := target.ValueOf("IpAddress")
		port, err := strconv.ParseUint(target.ValueOf("PortNumber"), 10, 16)
		if err != nil {
			return err
		}
		c.RelayTargetAddress = fmt.Sprintf("%s:%d", ip, port)
		c.SecuredRelayTarget, err = strconv.ParseBool(target.ValueOf("Secured"))
		if err != nil {
			return err
		}
		c.RelayTargetType = REDIRECT
	case "listening":
		target, err := configINI.Section("Listening")
		if err != nil {
			return err
		}
		port, err := strconv.ParseUint(target.ValueOf("PortNumber"), 10, 16)
		if err != nil {
			return err
		}
		c.RelayTargetAddress = fmt.Sprintf(":%d", port)
		c.SecuredRelayTarget, err = strconv.ParseBool(target.ValueOf("Secured"))
		if err != nil {
			return err
		}
		c.RelayTargetType = LISTENING
	default:
		return fmt.Errorf("unknown relay target type '%s', should be either normal or listen", relayTypeStr)
	}
	return nil
}

func SetupConfig(filePath string) (*Config, error) {
	var appConfig Config

	configINI, err := configparser.Read(filePath)
	if err != nil {
		return nil, err
	}
	server, err := configINI.Section("Server")
	if err != nil {
		return nil, err
	}
	appConfig.MiddleServerPort, err = strconv.Atoi(server.ValueOf("PortNumber"))
	if err != nil {
		return nil, err
	}
	appConfig.SecuredMiddleServer, err = strconv.ParseBool(server.ValueOf("Secured"))
	if err != nil {
		return nil, err
	}
	err = appConfig.setupRelayTargetConfig(server.ValueOf("RelayTargetType"), configINI)
	if err != nil {
		return nil, err
	}

	return &appConfig, nil
}
