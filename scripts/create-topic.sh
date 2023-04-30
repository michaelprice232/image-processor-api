#!/bin/zsh

topic=$1

docker compose exec broker \
  kafka-topics --create \
    --topic "${topic}" \
    --bootstrap-server localhost:9092 \
    --replication-factor 1 \
    --partitions 1