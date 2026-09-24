#!/usr/bin/env sh
set -eu

if [ "$#" -ne 2 ]; then
  echo "usage: $0 EVU TRAIN_NUMBER" >&2
  exit 1
fi

evu=$1
train_number=$2
operation_date=$(date +%F)

curl --silent --show-error \
  --get "http://localhost:3001/formation" \
  --data-urlencode "date=${operation_date}" \
  --data-urlencode "evu=${evu}" \
  --data-urlencode "trainNumber=${train_number}" \
  --output response.json

echo "Saved response to response.json"
