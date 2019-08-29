#!/bin/bash
loop=$1
while [ $loop -gt 0 ];do
	./main.exe --endpoint "ws://127.0.0.1:19736" send --from $2 --to $3 --token QLC --amount $4

	let loop--
	echo $loop
done
