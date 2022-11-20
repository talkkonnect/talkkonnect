#!/bin/bash
# Bash Script to Build murmur client version 1.5 from source on ubuntu desktop 
# By Suvir Kumar <suvir@talkkonnect.com>

apt install -y build-essential cmake pkg-config qttools5-dev qttools5-dev-tools libqt5svg5-dev libboost-dev
apt install -y libssl-dev libprotobuf-dev protobuf-compiler libprotoc-dev libcap-dev libxi-dev libasound2-dev
apt install -y libogg-dev libsndfile1-dev libspeechd-dev libavahi-compat-libdnssd-dev libxcb-xinerama0 libzeroc-ice-dev libpoco-dev g++-multilib

cd /usr/local/src
git clone https://github.com/mumble-voip/mumble
cd mumble
git submodule init
git submodule update
mkdir build
cd build
cd build/
cmake .. -DOPUS_CUSTOM_MODES=ON
make -j8
make install
