#!/usr/bin/env sh

# start processes
deliveryd start > ./logs/deliveryd.log &
deliveryd rest-server > ./logs/deliveryd-rest-server.log &
sleep 100
bridge start --all > ./logs/bridge.log &

# tail logs
tail -f ./logs/deliveryd.log ./logs/deliveryd-rest-server.log ./logs/bridge.log
