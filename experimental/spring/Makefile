build-example-docker-image:
	./mvnw -U clean install -P'os-linux,!os-darwin'
	cd spring-boot-nasp-example && ./mvnw -U spring-boot:build-image -P'os-linux,!os-darwin' -Dnasp.os=linux
