## TCP-Metadata Filter

This is a tcp-metadata filter written in Rust using proxy-wasm api.

### How to compile:

```bash
cargo +nightly build -Z build-std=std,panic_abort -Z build-std-features=panic_immediate_abort -j1 --release --target wasm32-unknown-unknown
```

To run this command your environment may require the following:

```bash
# Nightly toolchain must be installed
rustup toolchain install nightly
rustup component add rust-src --toolchain nightly-aarch64-apple-darwin
```

Optimalize the binary size of the created module even further by running:

```bash
wasm-opt -Os target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm -o target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm
```

## Stats filter

To build, first clone the repo of the wasm extensions of Istio:

```bash
git clone git@github.com:istio-ecosystem/wasm-extensions.git
cd wasm-extensions
```

Now start a Bazel container (use the specified image version, other versions simply don't work on Mac, and ARM:

```bash
docker run -v $PWD:/tmp/work -w /tmp/work --rm -it --entrypoint bash gcr.io/bazel-public/bazel:5.4.0
```

Inside the container run:

```bash
build '//extensions/stats:stats.wasm' # This takes some time if you are on an ARM Mac
cp bazel-bin/extensions/stats/stats.wasm .
```

You can find the stats.wasm now in your `wasm-extensions` working directory.
