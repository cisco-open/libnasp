include ../../common.mk

ios:
	gomobile bind -target ios -v -cache ${TMPDIR}/build/ios -o examples/ios-istio/Istio.xcframework	

android:
	gomobile bind -target android -v -androidapi 19 -cache ${TMPDIR}/build/android -o examples/android-istio/Istio.aar

mac:
	gomobile bind -target macos -v -cache ${TMPDIR}/build/mac -o examples/cli-istio/Istio.xcframework
