# v8

version: 5.1.281.75

# How to compile?

## 1. Download node.js

[node.js v6.9.1](https://nodejs.org/download/release/v6.9.1/)

## 2. Compile node.js

```sh
$ tar zxvf node-v6.9.1.tar.gz
$ cd node-v6.9.1
$ ./configure
$ make
```

## 3. Copy v8 library

```sh
$ find ./out -name *.a
./out/Release/obj.target/deps/gtest/libgtest.a
./out/Release/obj.target/deps/zlib/libzlib.a
./out/Release/obj.target/deps/http_parser/libhttp_parser.a
./out/Release/obj.target/deps/v8/tools/gyp/libv8_libbase.a
./out/Release/obj.target/deps/v8/tools/gyp/libv8_nosnapshot.a
./out/Release/obj.target/deps/v8/tools/gyp/libv8_base.a
./out/Release/obj.target/deps/v8/tools/gyp/libv8_libplatform.a
./out/Release/obj.target/deps/v8/tools/gyp/libv8_snapshot.a
./out/Release/obj.target/deps/cares/libcares.a
./out/Release/obj.target/deps/openssl/libopenssl.a
./out/Release/obj.target/deps/uv/libuv.a
./out/Release/obj.target/deps/v8_inspector/third_party/v8_inspector/platform/v8_inspector/libv8_inspector_stl.a
./out/Release/obj.target/tools/icu/libicudata.a
./out/Release/obj.target/tools/icu/libicuucx.a
./out/Release/obj.target/tools/icu/libicustubdata.a
./out/Release/obj.target/tools/icu/libicui18n.a
./out/Release/obj.host/tools/icu/libicutools.a

$ cp ./out/Release/obj.target/deps/v8/tools/gyp/*.a $GO_V8/lib/linux_amd64/
$ cp ./out/Release/obj.target/tools/icu/*.a $GO_V8/lib/linux_amd64/
```

## 4. Copy v8 header file

```sh
$ cp -r ./deps/v8/include/* $GO_V8/include/
```
