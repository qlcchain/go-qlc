#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

TAG="latest"

if [[ $TRAVIS_TAG ]]; then
    TAG=$TRAVIS_TAG
    docker tag qlcchain/go-qlc:latest qlcchain/go-qlc:$TAG
    docker push qlcchain/go-qlc:$TAG
fi

docker push qlcchain/go-qlc:latest
