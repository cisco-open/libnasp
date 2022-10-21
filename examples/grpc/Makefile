proto:
	protoc --go_out=./pb --go_opt=paths=source_relative \
	--go-grpc_out=./pb --go-grpc_opt=paths=source_relative \
	hello.proto

# brew install libressl
cert:
	/usr/local/opt/libressl/bin/openssl req -x509 \
	-newkey rsa:4096 -keyout key.pem -out cert.pem \
	-sha256 -days 365 -nodes \
	-addext "subjectAltName = DNS:localhost" \
	-subj "/C=HU/ST=Budapest/L=Budapest/O=Cisco/CN=localhost"

curl:
	grpcurl -plaintext -d '{"name":"Joska"}' localhost:1234 Greeter/SayHello

run-server:
	go run .

run-client:
	go run ./client
