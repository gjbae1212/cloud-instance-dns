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
    go build
}

function linux_build
{
   GOOS=linux GOARCH=amd64 go build -o dist/cloud-instance-dns
}

CMD=$1
shift
$CMD $*