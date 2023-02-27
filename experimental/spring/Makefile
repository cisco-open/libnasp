build-example-docker-image:
	rm -rf ~/.m2/repository/nasp/nasp/0.0.0/
	./mvnw -U clean install -P'linux,!darwin'
	cd spring-boot-nasp-example && ./mvnw -U jib:dockerBuild -P'linux,!darwin' -Dnasp.os=linux
