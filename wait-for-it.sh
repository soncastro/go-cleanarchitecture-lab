#!/usr/bin/env bash
#   Use this script to test if a given TCP host/port are available

# The MIT License (MIT)
# Copyright (c) 2016 Vincent Ambo
# Full license is provided in the 'wait-for-it.sh' source code file

set -e

TIMEOUT=15
QUIET=0

while getopts ":t:q" opt; do
  case ${opt} in
    t)
      TIMEOUT=$OPTARG
      ;;
    q)
      QUIET=1
      ;;
    \?)
      echo "Invalid option: $OPTARG" 1>&2
      exit 1
      ;;
    :)
      echo "Invalid option: $OPTARG requires an argument" 1>&2
      exit 1
      ;;
  esac
done
shift $((OPTIND -1))

if [ $# -ne 1 ]; then
  echo "Usage: $0 host:port" 1>&2
  exit 1
fi

HOST_PORT=$1
HOST=$(echo $HOST_PORT | cut -d : -f 1)
PORT=$(echo $HOST_PORT | cut -d : -f 2)

for i in $(seq $TIMEOUT) ; do
  if nc -z $HOST $PORT ; then
    if [ $QUIET -ne 1 ]; then
      echo "$HOST:$PORT is available after $i seconds"
    fi
    exit 0
  fi
  sleep 1
done

echo "Timeout occurred after waiting $TIMEOUT seconds for $HOST:$PORT" 1>&2
exit 1
