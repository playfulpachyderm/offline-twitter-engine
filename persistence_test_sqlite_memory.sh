#!/bin/bash

sudo mount -t tmpfs -o size=100M tmpfs pkg/persistence/test_profiles
sudo mount -t tmpfs -o size=500M tmpfs cmd/data
sudo mount -t tmpfs -o size=1000M tmpfs sample_data/profile
