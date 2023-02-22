build-example-docker-image:
	./mvnw -U clean install -Pos-linux
	cd spring-boot-nasp-example && ./mvnw -U spring-boot:build-image -Pos-linux
