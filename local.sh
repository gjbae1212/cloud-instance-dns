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

CMD=$1
shift
$CMD $*
