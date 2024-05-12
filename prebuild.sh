#!/bin/bash

set -x

script_dir="$(cd "$(dirname "$0")"; pwd)"
cd "$script_dir"

rm -rf ./src/files ./.tmp ./tmp ./*.tar ./src/*.tar ./rsync.exe.stackdump 
mkdir -p .tmp

cp -rf ./files ./makefile ./prebuild.sh ./readme.md ./src "./test data" ./.tmp

# go:embed で埋め込むためにpackage mainのあるパスに移動させる。
cd ./.tmp
tar cvf ../embed.tar *
cd ..
rm -rf ./.tmp

# go:embed で埋め込むためにpackage mainのあるパスに移動させる。
mv ./embed.tar src/

