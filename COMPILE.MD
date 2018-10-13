How to build? First you will need some libraries 
* install required librarys: LibOpenAl, LibAsound
* on raspbian apt-get install libasound-dev libopenal-dev
* export the GOPATH & GOBIN directorys using env variables:
* export GOPATH=/home/mumble/go
* export GOBIN=/home/mumble/bin
* go get github.com/talkkonnect/talkkonnect
* cd $GOPATH/src/github.com/talkkonnect/talkkonnect
* go build -o /home/mumble/bin/talkkonnect cmd/talkkonnect/main.go
