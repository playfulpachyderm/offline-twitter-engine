#!/bin/bash

if [[ ! -f sample_data/profile/Offline_Twatter.session ]]; then
	cp ~/twitter/*.session sample_data/profile/
fi
go run ./cmd/twitter --profile sample_data/profile --session Offline_Twatter webserver --addr localhost:1487 --auto-open
