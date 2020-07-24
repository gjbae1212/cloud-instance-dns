#!/bin/bash
set -e -o pipefail
trap '[ "$?" -eq 0 ] || echo "Error Line:<$LINENO> Error Function:<${FUNCNAME}>"' EXIT

cd `dirname $0`
CURRENT=`pwd`

function test
{
   setenv
   go test -v $(go list ./... | grep -v vendor) --count 1 -race -coverprofile=$CURRENT/coverage.txt -covermode=atomic

}

function setenv
{
     if [ -e $CURRENT/local_env.sh ]; then
         source $CURRENT/local_env.sh
     fi
}

function build
{
    cd $CURRENT/cmd/dns
    go build -trimpath -ldflags='-s -w' -o $CURRENT/dist/dns-darwin
}

function linux_build
{
   cd $CURRENT/cmd/dns
   GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o $CURRENT/dist/dns-linux
}

function release
{
  # export GITHUB_TOKEN=blahblah
  # before --> git tag -a 'version' -m ''
  cd $CURRENT/cmd/dns
  goreleaser release --rm-dist
}

CMD=$1
shift
$CMD $*
