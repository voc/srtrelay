#!/bin/sh
set -e
sudo apt-get install -y tclsh pkg-config cmake libssl-dev build-essential ffmpeg ninja-build
git clone --depth 1 --branch v1.5.0 https://github.com/Haivision/srt.git libsrt
cmake -S libsrt -B libsrt-build -G Ninja
ninja -C libsrt-build
ninja -C libsrt-build install
