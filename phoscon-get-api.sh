#!/bin/bash

set -eu

if [ $# -lt 1 ]; then
    echo "Usage: $0 <PHOSCON API HOST>"
    exit 1
fi

echo "Go to Menu/Settings/Gateway/Advanced and unlock clicking" \
    "on \"Authenticate app\", then press any key to continue or ctrl-c to abort"
read -n 1

curl -d "{\"devicetype\": \"$(hostname)\"}" -H "Content-Type: application/json" -X POST "$1/api"

echo -e "Done. If you received a successful response, write down the \"username\" value and use it as API Token"