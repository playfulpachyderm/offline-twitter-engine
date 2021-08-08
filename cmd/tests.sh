#!/bin/bash

set -e
set -x

PS4='+(${BASH_SOURCE}:${LINENO}): '

./compile.sh

test -e data && rm -r data

./tw create_profile data

# Fetch a user
./tw fetch_user data Denlesks
test $(sqlite3 data/twitter.db "select handle from users") = "Denlesks"
test $(sqlite3 data/twitter.db "select count(*) from users") = "1"
./tw fetch_user data Denlesks
test $(sqlite3 data/twitter.db "select count(*) from users") = "1"

# Fetch a tweet with images
./tw fetch_tweet_only data https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "1"
test "$(sqlite3 data/twitter.db "select text from tweets")" = "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do"
test $(sqlite3 data/twitter.db "select count(*) from images") = "4"

# Download its images
test $(sqlite3 data/twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "4"
test $(sqlite3 data/twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "0"
test $(sqlite3 data/twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "0"
test $(find data/images | wc -l) = "1"
./tw download_tweet_content data 1261483383483293700
test $(sqlite3 data/twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "0"
test $(sqlite3 data/twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "4"
test $(sqlite3 data/twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "1"
test $(find data/images | wc -l) = "5"

# Try to double-download it
./tw fetch_tweet_only data https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "1"
test $(sqlite3 data/twitter.db "select count(*) from images") = "4"


# Fetch a tweet with a video
./tw fetch_user data DiamondChariots
test $(sqlite3 data/twitter.db "select handle from users" | wc -l) = "2"
./tw fetch_tweet_only data https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 data/twitter.db "select count(*) from videos") = "1"

# Download the video
test $(sqlite3 data/twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "1"
test $(sqlite3 data/twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "0"
test $(sqlite3 data/twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "0"
test $(find data/videos| wc -l) = "1"
./tw download_tweet_content data 1418971605674467340
test $(sqlite3 data/twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "0"
test $(sqlite3 data/twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "1"
test $(sqlite3 data/twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "1"
test $(find data/videos | wc -l) = "2"

# Try to double-download it
./tw fetch_tweet_only data https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 data/twitter.db "select count(*) from videos") = "1"


echo -e "\033[32mAll tests passed.  Finished successfully.\033[0m"
