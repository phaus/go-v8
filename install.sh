#!/bin/bash

# build v8 native version
cd v8
svn co http://gyp.googlecode.com/svn/trunk build/gyp --revision 1831
make i18nsupport=off native
cd ..

outdir="`pwd`/v8/out/native"

libv8_base="`find $outdir -name 'libv8_base.*.a' | head -1`"
if [ ! -f $libv8_base ]; then
	echo >&2 "V8 build failed?"
	exit
fi

# for Linux
librt=''
if [ `go env | grep GOHOSTOS` == 'GOHOSTOS="linux"' ]; then
	librt='-lrt'
fi

# for Mac
libstdcpp=''
if  [ `go env | grep GOHOSTOS` == 'GOHOSTOS="darwin"' ]; then
	libstdcpp='-stdlib=libstdc++'
fi

# create package config file
echo "Name: v8
Description: v8 javascript engine
Version: $v8_version
Cflags: $libstdcpp -I`pwd` -I`pwd`/v8/include
Libs: $libstdcpp $libv8_base $outdir/libv8_snapshot.a $librt" > v8.pc

# let's go
go install
go test -v