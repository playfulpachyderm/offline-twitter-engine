#!/bin/bash

set -e

THIS_DIR=$(readlink -f $0 | xargs dirname)

if [[ -d "$THIS_DIR/profile" ]] && [ "$(ls -A "$THIS_DIR"/profile)" ]; then
	rm -r $THIS_DIR/profile/*
fi
mkdir -p $THIS_DIR/profile

test -e $THIS_DIR/profile/twitter.db && rm $THIS_DIR/profile/twitter.db
sqlite3 $THIS_DIR/profile/twitter.db < $THIS_DIR/seed_data.sql

mkdir $THIS_DIR/profile/profile_images
cp $THIS_DIR/kwamurai_* $THIS_DIR/profile/profile_images
cp $THIS_DIR/../pkg/persistence/default_profile.png $THIS_DIR/profile/profile_images

mkdir $THIS_DIR/profile/images
cp $THIS_DIR/EYG* $THIS_DIR/profile/images

mkdir $THIS_DIR/profile/videos
cp $THIS_DIR/*.mp4 $THIS_DIR/profile/videos

mkdir $THIS_DIR/profile/link_preview_images
mkdir $THIS_DIR/profile/video_thumbnails
