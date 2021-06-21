package config

import "crypto/tls"

func ServerTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func ClientTLSConfig() (*tls.Config, error) {
	return &tls.Config{InsecureSkipVerify: true}, nil
}