package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"tcprelay"
	"tcprelay/config"
	"tcprelay/relaytarget"
)

var (
	serverTLSConfig *tls.Config
	clientTLSConfig *tls.Config
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	cfg, err := config.SetupConfig("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("starting server with configurations below: \n%s", cfg)

	if cfg.SecuredMiddleServer {
		serverTLSConfig, err = config.ServerTLSConfig("certs/server.cert", "cert/server.key")
		if err != nil {
			log.Fatal(err)
		}
	}
	var target relaytarget.TcpRelayTarget
	if cfg.RelayTargetType == config.LISTENING {
		var targetTLSConfig *tls.Config
		if cfg.SecuredRelayTarget {
			targetTLSConfig = serverTLSConfig
		}
		target = relaytarget.NewListenableRelayTarget(cfg.RelayTargetAddress, targetTLSConfig)
	} else if cfg.RelayTargetType == config.REDIRECT {
		var targetTLSConfig *tls.Config
		if cfg.SecuredRelayTarget {
			targetTLSConfig, err = config.ClientTLSConfig()
			if err != nil {
				log.Fatal(err)
			}
		}
		target = relaytarget.NewRelayTarget(cfg.RelayTargetAddress, targetTLSConfig)
	}

	s := tcprelay.NewTcpRelayServer(
		cfg.MiddleServerPort,
		target,
		serverTLSConfig,
	)
	go s.Listen()

	done := make(chan bool, 1)
	<-done
}
