
```bash
./mvnw spring-boot:build-image -Pos-linux
kind load docker-image --name nasp-test-cluster docker.io/library/spring-boot-nasp-example:1.0-SNAPSHOT
kubectl create deployment --image docker.io/library/spring-boot-nasp-example:1.0-SNAPSHOT spring-boot-nasp
```
