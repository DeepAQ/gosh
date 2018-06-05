#!/bin/bash

ETCD_HOST=etcd
ETCD_PORT=2379
ETCD_URL=http://$ETCD_HOST:$ETCD_PORT

echo ETCD_URL = $ETCD_URL

if [[ "$1" == "consumer" ]]; then
  echo "Starting consumer agent..."
  gosh -type=consumer -port=20000 -etcd=$ETCD_URL
elif [[ "$1" == "provider-small" ]]; then
  echo "Starting small provider agent..."
  gosh -type=provider -weight=10 -port=30000 -dubbo.port=20880 -etcd=$ETCD_URL
elif [[ "$1" == "provider-medium" ]]; then
  echo "Starting medium provider agent..."
  gosh -type=provider -weight=40 -port=30000 -dubbo.port=20880 -etcd=$ETCD_URL
elif [[ "$1" == "provider-large" ]]; then
  echo "Starting large provider agent..."
  gosh -type=provider -weight=50 -port=30000 -dubbo.port=20880 -etcd=$ETCD_URL
else
  echo "Unrecognized arguments, exit."
  exit 1
fi