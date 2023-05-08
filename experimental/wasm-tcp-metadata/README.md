TCP-Metadata Filter

This is a tcp-metadata filter written in Rust using proxy-wasm api.

How to compile:

```
cargo +nightly build -Z build-std=std,panic_abort -Z build-std-features=panic_immediate_abort -j1 --release --target wasm32-unknown-unknown
```

To run this command your environment may require the following:

```

# Nightly toolchain must be installed
rustup toolchain install nightly
rustup component add rust-src --toolchain nightly-aarch64-apple-darwin
```

Optimalize the binary size of the created module even further by running:
```
wasm-opt -Os target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm -o target/wasm32-unknown-unknown/release/wasm_tcp_metadata.wasm
```
