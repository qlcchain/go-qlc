#!/bin/bash

network='live'

print_usage() {
	echo 'build.sh [-h] [-n {live|beta|test}]'
}

while getopts 'hn:' OPT; do
	case "${OPT}" in
		h)
			print_usage
			exit 0
			;;
		n)
			network="${OPTARG}"
			;;
		*)
			print_usage >&2
			exit 1
			;;
	esac
done

case "${network}" in
	live)
		network_tag=''
		build_flag='build'
		;;
	test|beta)
		network_tag="-test"
		build_flag='build-test'
		;;
	*)
		echo "Invalid network: ${network}" >&2
		exit 1
		;;
esac

REPO_ROOT=`git rev-parse --show-toplevel`
COMMIT_SHA=`git rev-parse --short HEAD`
pushd $REPO_ROOT
echo ${build_flag}
docker build --build-arg BUILD_ACT="${build_flag}" -f docker/Dockerfile -t qlcchain/go-qlc${network_tag}:latest .
popd
