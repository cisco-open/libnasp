include ../../common.mk

ios:
	gomobile bind -target ios/arm64 -v -o examples/ios-istio/Istio.xcframework

android:
	gomobile bind -target android/arm64 -androidapi 19 -v -o examples/android-istio/Istio.aar

mac:
	gomobile bind -target macos -v -o examples/cli-istio/Istio.xcframework
