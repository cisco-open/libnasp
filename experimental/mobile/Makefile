include ../../common.mk

ios:
	gomobile bind -target ios/arm64,iossimulator/arm64 -v -o examples/ios-nasp/Nasp.xcframework

android:
	gomobile bind -target android/arm64 -androidapi 19 -v -o examples/android-nasp/lib/nasp.aar

mac:
	gomobile bind -target macos -v -o examples/cli-nasp/nasp.xcframework
