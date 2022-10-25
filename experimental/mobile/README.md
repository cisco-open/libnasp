# NASP on mobile

NASP is written mostly in Go - and some parts in other languages, but compiled to web assembly. Thanks to this, we can leverage [Go mobile](https://pkg.go.dev/golang.org/x/mobile) to compile the library to the most well-known mobile platforms to languages like Java/Kotlin (Android) and Objective-C/Swift (iOS) with [gobind](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind). Not to mention that Java and Swift bindings allows us to excercise these bindings in server-side applications as well.

If you are interested how the bindings work in the background there is a [great blogpost](https://medium.com/@matryer/tutorial-calling-go-code-from-swift-on-ios-and-vice-versa-with-gomobile-7925620c17a4) describing the internals.

This small library exposes only basic HTTP request capabilities and not yet complete, there is more to come.

## Compile

It's worth to have a look at the official Go mobile wiki to have the initatial understanding how it works: https://github.com/golang/go/wiki/Mobile 

### Prepare gomobile first

First we need to setup `gomobile` and `gobind` from the Go mobile project:

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
```

### Compile the library to a specific mobile platform library

### iOS/XCFramework

Xcode installation is neccessary.

```bash
cd experimental/mobile
make ios
```

### Android/AAR

Android SDK/NDK installation is neccessary.

```bash
cd experimental/mobile
make android
```

### Example applications

There are two basic mobile applications in the `expermental/mobile/examples` directory for each aforementioned platform.
It is a basic application, which can connect to an existing Istio cluster via getting the initial Istio configuration through [Heimdall](../heimdall).
