# Build

To be able to build the required nasp JAR library we need to utilize a forked version of Go Mobile temporarily:

```bash
git clone -b plainjava git@github.com:bonifaido/mobile.git
cd mobile
go install ./cmd/gomobile
go install ./cmd/gobind
```

In the NASP project root:

```
cd experimental/mobile
make java
```

