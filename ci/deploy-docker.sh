#!/bin/bash
set -e

scripts="$(dirname "$0")"

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

tags=()
if [[ -n "$TRAVIS_TAG" ]]; then
    tags+=("$TRAVIS_TAG" latest)
elif [[ -n "$TRAVIS_BRANCH" ]]; then
    if [[ "$TRAVIS_BRANCH" != "master" ]]; then
        tags+=("$TRAVIS_BRANCH")
    fi
fi

for network in live test; do
    if [ "${network}" = 'live' ]; then
        network_tag_suffix=''
        build_flag='build'
    else
        network_tag_suffix="-${network}"
        build_flag='build-test'
    fi

    docker_image_name="qlcchain/go-qlc${network_tag_suffix}"
    echo "build ${network} ==> ${network_tag_suffix} ${build_flag}"
    "$scripts"/custom-timeout.sh 30 docker build --build-arg BUILD_ACT=${build_flag} -f docker/Dockerfile -t "$docker_image_name" .
    for tag in "${tags[@]}"; do
        # Sanitize docker tag
        # https://docs.docker.com/engine/reference/commandline/tag/
        tag="$(printf '%s' "$tag" | tr -c '[a-z][A-Z][0-9]_.-' -)"
        if [ "$tag" != "latest" ]; then
            docker tag "$docker_image_name" "${docker_image_name}:$tag"
        fi
        "$scripts"/custom-timeout.sh 30 docker push "${docker_image_name}:$tag"
    done
done
