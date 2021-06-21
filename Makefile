server:
	go run cmd/server/main.go

client:
	go run cmd/client/main.go

echo-client:
	go run cmd/echo/main.go connect -ip $(IP) -port $(PORT)

echo-server:
	go run cmd/echo/main.go listen -port $(PORT)

certs:
	mkdir -p certs/
	openssl genrsa -out certs/server.key 2048
	openssl req -new -x509 -sha256 -key certs/server.key -out certs/server.crt -days 3650

build:
	go build -o bin/server.exe cmd/server/main.go
	cp config.ini bin/config.ini

clean:
	rm -rf bin/

.PHONY: server client echo-client certs build clean