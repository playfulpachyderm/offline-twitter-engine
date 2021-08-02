#!/bin/bash

set -e
set -x

test -e data && rm -r data

go run ./twitter create_profile data

# Fetch a user
go run ./twitter fetch_user data Denlesks
test $(sqlite3 data/twitter.db "select handle from users") = "Denlesks"
test $(sqlite3 data/twitter.db "select count(*) from users") = "1"
go run ./twitter fetch_user data Denlesks
test $(sqlite3 data/twitter.db "select count(*) from users") = "1"

# Fetch a tweet with images
go run ./twitter fetch_tweet_only data https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "1"
test "$(sqlite3 data/twitter.db "select text from tweets")" = "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do"
go run ./twitter fetch_tweet_only data https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 data/twitter.db "select count(*) from tweets") = "1"
