#!/usr/bin/env bash
echo "start qlc node"

params=()

function join_by() {
    local IFS="$1"
    shift
    echo "$*"
}

if [[ -n "$QLC_SEED" ]]; then
    echo "run as seed"
    params+=("--seed=${QLC_SEED}")
elif [[ -n "${seed}" ]]; then
    echo "run as seed(old)"
    params+=("--seed=${seed}")
elif [[ -n "${QLC_ACCOUNT}" ]]; then
    echo "run as account(private)"
    params+=("--privateKey=${QLC_ACCOUNT}")
fi

if [[ -n "${QLC_PARAMS}" ]]; then
    echo "set configParams ${QLC_PARAMS}"
    params+=("--configParams=${QLC_PARAMS}")
fi

if [[ -n "${QLC_NO_BOOTNODE}" ]]; then
    echo 'disable all default bootnodes'
    params+=("--nobootnode=true")
fi

result="$(join_by ' ' ${params[@]})"
echo "params: ${result}"

./gqlc "${result}" && echo 'start go-qlc successful'
