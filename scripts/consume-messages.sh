#!/bin/zsh

topic=$1
docker compose exec -it broker kafka-console-consumer --bootstrap-server localhost:9092 --topic "${topic}"