#!/bin/bash
# Bash Script to Build murmur server version 1.5 from source on debian 11 minimal OS 
# By Suvir Kumar <suvir@talkkonnect.com>

apt install -y build-essential pkg-config qttools5-dev qttools5-dev-tools libqt5svg5-dev libboost-dev libssl-dev libprotobuf-dev protobuf-compiler 
apt install -y libprotoc-dev libcap-dev libxi-dev libasound2-dev libogg-dev
apt install -y libsndfile1-dev libspeechd-dev libavahi-compat-libdnssd-dev libxcb-xinerama0 libzeroc-ice-dev libpoco-dev git
apt install -y qtcreator
apt install -y libcurl4-openssl-dev libavahi-compat-libdnssd-dev libssl-dev libzeroc-ice-dev
apt install -y libboost-tools-dev libboost-thread-dev magics++
apt install -y net-tools wget

cd /usr/src
git clone https://github.com/Kitware/CMake

cd /usr/src/CMake
./bootstrap
make
make install

cd /usr/src/
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.20.0/protobuf-all-3.20.0.tar.gz
tar -zxvf protobuf-all-3.20.0.tar.gz
cd /usr/src/protobuf-3.20.0/
./configure
make
make install
ldconfig
protoc --version

cd /usr/src
git clone https://github.com/mumble-voip/mumble/
cd /usr/src/mumble
git submodule update --init --recursive

cd /usr/src/mumble
mkdir build
cd /usr/src/mumble/build
cmake -Dclient=OFF ..
make
make install

exit 0


