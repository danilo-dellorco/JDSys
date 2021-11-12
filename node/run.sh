#!/bin/sh
cd /home/ec2-user/go/src/progetto-sdcc
git pull
sleep 10
cd node/main
sudo go run node.go 10.0.0.64 > /home/ec2-user/log/progetto-sdcc.log