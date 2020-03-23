#!/bin/bash

readonly image="$1"
readonly tag="$2"

echo "$3" > ./print_this.txt


docker build -t ${image}:${tag} .
docker push ${image}:${tag}

