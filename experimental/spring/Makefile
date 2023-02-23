build-example-docker-image:
	rm -rf ~/.m2/repository/nasp/nasp/0.0.0/
	./mvnw -U clean install -P'linux,!darwin' -P'linux,!darwin'
	./mvnw -U -pl spring-boot-nasp-example jib:dockerBuild -P'linux,!darwin' -Dnasp.os=linux
