#!/bin/sh
set -e
sudo apt-get install -y tclsh pkg-config cmake libssl-dev build-essential ffmpeg
git clone --depth 1 --branch v1.4.2 https://github.com/Haivision/srt.git libsrt
cd libsrt
./configure
make -j
sudo make install
