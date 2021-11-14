#!/bin/sh
cd /home/ec2-user/go/src/progetto-sdcc
git pull
sleep 10
cd registry/main
sudo go run registry.go > /home/ec2-user/log/progetto-sdcc.log