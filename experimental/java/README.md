# NASP java library

## Build

To be able to build the required nasp JAR library we need to utilize a forked version of Go Mobile temporarily:

```bash
make install-gomobile
```

Build the glue code using gomobile:

```
make java
```

## Release

Artifacts are stored at the GitHub packages repo of the project, you need a GitHub Access token to be able to release.
See the official docs of [Maven registry for GitHub](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-apache-maven-registry).

```bash
mvn deploy
```
