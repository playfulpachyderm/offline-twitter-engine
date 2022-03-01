#!/bin/bash

set -e
set -x

PS4='+(${BASH_SOURCE}:${LINENO}): '

FAKE_VERSION="1.100.3489"
./compile.sh $FAKE_VERSION

test -e data && rm -r data

PATH=`pwd`:$PATH

test "$(tw --version)" = "v$FAKE_VERSION"

tw --help
test $? -eq 0

tw create_profile data
cd data


# Fetch a user
test $(find profile_images | wc -l) = "1"                      # should be empty to begin
tw fetch_user Denlesks
test "$(sqlite3 twitter.db "select handle from users")" = "Denlesks"
test $(sqlite3 twitter.db "select count(*) from users") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'Denlesks'") = "1"
test $(find profile_images | wc -l) = "3"                      # should have gotten 2 images
test -f profile_images/Denlesks_profile_22YJvhC7.jpg
test -f profile_images/Denlesks_banner_1585776052.jpg
tw fetch_user Denlesks                                         # try to double-download it
test $(sqlite3 twitter.db "select count(*) from users") = "1"  # shouldn't have added a new row


# Fetch a tweet with images
tw fetch_tweet_only https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test "$(sqlite3 twitter.db "select text from tweets")" = "These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do"
test $(sqlite3 twitter.db "select count(*) from images") = "4"

# Download its images
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "4"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "0"
test $(find images | wc -l) = "1"
tw download_tweet_content https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from images where tweet_id = 1261483383483293700 and is_downloaded = 1") = "4"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1261483383483293700") = "1"
test $(find images | wc -l) = "5"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/Denlesks/status/1261483383483293700
test $(sqlite3 twitter.db "select count(*) from tweets") = "1"
test $(sqlite3 twitter.db "select count(*) from images") = "4"


# Fetch a tweet with a video
tw fetch_user DiamondChariots
test $(sqlite3 twitter.db "select handle from users" | wc -l) = "2"
tw fetch_tweet_only https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"

# Download the video
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "1"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "0"
test $(find videos | wc -l) = "1"
test $(find video_thumbnails | wc -l) = "1"
tw download_tweet_content 1418971605674467340
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 0") = "0"
test $(sqlite3 twitter.db "select count(*) from videos where tweet_id = 1418971605674467340 and is_downloaded = 1") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1418971605674467340") = "1"
test $(find videos | wc -l) = "2"
test $(find video_thumbnails | wc -l) = "2"

# Try to double-download it
tw fetch_tweet_only https://twitter.com/DiamondChariots/status/1418971605674467340
test $(sqlite3 twitter.db "select count(*) from tweets") = "2"
test $(sqlite3 twitter.db "select count(*) from videos") = "1"


# Fetch a tweet with a GIF
tw fetch_user Cernovich
initial_videos_count=$(find videos | wc -l)
initial_videos_db_count=$(sqlite3 twitter.db "select count(*) from videos")
tw fetch_tweet_only https://twitter.com/Cernovich/status/1444429517020274693

test $(sqlite3 twitter.db "select count(*) from videos") = "$((initial_videos_db_count + 1))"
test $(sqlite3 twitter.db "select is_gif from videos where tweet_id = 1444429517020274693") = "1"

# Download the GIF
test $(find videos | wc -l) = "$((initial_videos_count))"   # Shouldn't have changed yet
tw download_tweet_content https://twitter.com/Cernovich/status/1444429517020274693
test $(find videos | wc -l) = "$((initial_videos_count + 1))"



# Download a full thread
test ! -e profile_images/default_profile.png
tw fetch_tweet https://twitter.com/RememberAfghan1/status/1429585423702052867
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429585423702052867") = "RememberAfghan1"
test $(sqlite3 twitter.db "select is_conversation_scraped, abs(last_scraped_at - strftime('%s','now')) < 30 from tweets where id = 1429585423702052867") = "1|1"
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429584239570391042") = "michaelmalice"
test $(sqlite3 twitter.db "select is_conversation_scraped from tweets where id = 1429584239570391042") = "0"
test "$(sqlite3 twitter.db "select handle, is_banned from tweets join users on tweets.user_id = users.id where tweets.id=1429583672827465730")" = "kanesays23|1"  # This guy got banned
test -e profile_images/default_profile.png
test $(sqlite3 twitter.db "select handle from tweets join users on tweets.user_id = users.id where tweets.id=1429616911315345414") = "NovaValentis"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429585423702052867") = "michaelmalice"
test $(sqlite3 twitter.db "select reply_mentions from tweets where id = 1429616911315345414") = "RememberAfghan1,michaelmalice"


# Test that profile images (tiny vs regular) are chosen properly
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'Cernovich'") = "1"
test $(find profile_images/Cernovich* | grep normal | wc -l) = "0"  # Since "Cernovich" was fetched directly, should have full-sized profile image and banner
test $(find profile_images/Cernovich* | grep banner | wc -l) = "1"

test $(sqlite3 twitter.db "select is_content_downloaded from users where handle = 'RememberAfghan1'") = "0"
test $(find profile_images/RememberAfghan1* | grep normal | wc -l) = "1"  # "RememberAfghan1" was fetched via a tweet thread and isn't followed, so should have tiny profile image and no banner
test $(find profile_images/RememberAfghan1* | grep banner | wc -l) = "0"


# Test that the `--profile` flag works
cd ..
tw --profile data fetch_user elonmusk
test $(sqlite3 data/twitter.db "select count(*) from users where handle = 'elonmusk'") = "1"
cd data

# Test that fetching tweets with ID only (not full URL) works
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "0"  # Check it's not already there
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='elonmusk'") = "1"  # Should be downloaded from the previous test!
tw fetch_tweet 1433713164546293767
test $(sqlite3 twitter.db "select count(*) from tweets where id = 1433713164546293767") = "1"  # Should be there now
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='elonmusk'") = "1"  # Should not un-set content-downloaded!


# Get a user's feed
malice_id=$(sqlite3 twitter.db "select id from users where handle='michaelmalice'")
test $(sqlite3 twitter.db "select count(*) from retweets") = "0"
tweet_count_1=$(sqlite3 twitter.db "select count(*) from tweets")
tw get_user_tweets michaelmalice

# Check that there are some retweets
rts_count=$(sqlite3 twitter.db "select count(*) from retweets")
test $rts_count -gt "0"

# Check that new retweets plus new tweets > 50
tweet_count_2=$(sqlite3 twitter.db "select count(*) from tweets")
test $(sqlite3 twitter.db "select count(*) from retweets where retweeted_by != $malice_id") = "0"
test $(($rts_count + $tweet_count_2 - $tweet_count_1)) -gt "50"


# Fetch a privated user
tw fetch_user LandsharkRides
test $(sqlite3 twitter.db "select is_private from users where handle = 'LandsharkRides'") = "1"


# Test tweets with URLs
tw fetch_user CovfefeAnon
urls_count=$(sqlite3 twitter.db "select count(*) from urls")
tw fetch_tweet_only https://twitter.com/CovfefeAnon/status/1428904664645394433
urls_count_after=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after = $(($urls_count + 1))
test "$(sqlite3 twitter.db "select title from urls where tweet_id = 1428904664645394433")" = "Justice Department investigating Elon Musk's SpaceX following complaint of hiring discrimination"
test $(sqlite3 twitter.db "select count(*) from urls where tweet_id = 1428904664645394433") = "1"
thumbnail_name=$(sqlite3 twitter.db "select thumbnail_remote_url from urls where tweet_id = 1428904664645394433" | grep -Po "(?<=/)[\w-]+(?=\?)")
test -n "$thumbnail_name"  # Not testing for what the thumbnail url is because it keeps changing

# Try to double-fetch it; shouldn't duplicate the URL
tw fetch_tweet_only https://twitter.com/CovfefeAnon/status/1428904664645394433
urls_count_after_2x=$(sqlite3 twitter.db "select count(*) from urls")
test $urls_count_after_2x = $urls_count_after

# Download the link's preview image
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1428904664645394433") = "0"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1428904664645394433") = "0"
initial_link_preview_images_count=$(find link_preview_images | wc -l)
tw download_tweet_content 1428904664645394433
test $(sqlite3 twitter.db "select is_content_downloaded from tweets where id = 1428904664645394433") = "1"
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1428904664645394433") = "1"
test $(find link_preview_images | wc -l) = "$((initial_link_preview_images_count + 1))"
test -f link_preview_images/${thumbnail_name}_800x320_1.jpg


# Test a tweet with a URL but no thumbnail
tw fetch_user Xirong7
tw fetch_tweet_only https://twitter.com/Xirong7/status/1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "0"
test $(sqlite3 twitter.db "select has_thumbnail from urls where tweet_id = 1413665734866186243") = "0"
initial_link_preview_images_count=$(find link_preview_images | wc -l)  # Check that it doesn't change, since there's no thumbnail
tw download_tweet_content 1413665734866186243
test $(sqlite3 twitter.db "select is_content_downloaded from urls where tweet_id = 1413665734866186243") = "1"
test $(find link_preview_images | wc -l) = $initial_link_preview_images_count  # Should be the same


# Test a tweet thread with tombstones
# tw fetch_tweet https://twitter.com/CovfefeAnon/status/1454526270809726977
# test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454515503242829830") = 1
# test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454521424144654344") = 0  # TODO this guy got banned
# test $(sqlite3 twitter.db "select is_stub from tweets where id = 1454522147750260742") = 1


# Test a tweet thread with a deleted account; should generate a user with a fake ID
tw fetch_tweet https://twitter.com/michaelmalice/status/1497272681497980928
test $(sqlite3 twitter.db "select is_id_fake from users where handle = 'GregCunningham0'") = 1
test $(sqlite3 twitter.db "select count(*) from tweets where user_id = (select id from users where handle = 'GregCunningham0')") = 1


# Test search
tw search "from:michaelmalice constitution"
test $(sqlite3 twitter.db "select count(*) from tweets where user_id = 44067298 and text like '%constitution%'") -gt "30"  # Not sure exactly how many

tw fetch_tweet 1465534109573390348
test $(sqlite3 twitter.db "select count(*) from polls where tweet_id = 1465534109573390348") = "1"
test "$(sqlite3 twitter.db "select choice1, choice2, choice3, choice4 from polls where tweet_id = 1465534109573390348")" = "Tribal armband|Marijuana leaf|Butterfly|Maple leaf"
test "$(sqlite3 twitter.db "select choice1_votes, choice2_votes, choice3_votes, choice4_votes from polls where tweet_id = 1465534109573390348")" = "1593|624|778|1138"


# Test fetching a banned user
rm profile_images/default_profile.png
tw fetch_user nancytracker
test "$(sqlite3 twitter.db "select is_content_downloaded, is_banned from users where handle='nancytracker'")" = "1|1"
test -e profile_images/default_profile.png


# Fetch a user with "600x200" banner image
tw fetch_user AlexKoppelman  # This is probably kind of a flimsy test
test $(sqlite3 twitter.db "select is_content_downloaded from users where handle='AlexKoppelman'") = "1"


# Test following / unfollowing a user
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "0"
tw follow michaelmalice
test "$(sqlite3 twitter.db "select handle from users where is_followed = 1")" = "michaelmalice"

tw follow cernovich
test "$(tw list_followed | wc -l)" = 2
test "$(tw list_followed | grep -iq cernovich && echo YES)" = "YES"
test "$(tw list_followed | grep -iq michaelmalice && echo YES)" = "YES"
test "$(tw list_followed | grep -iq blahblahgibberish && echo YES)" = ""

tw unfollow michaelmalice
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "1"
tw unfollow cernovich
test "$(sqlite3 twitter.db "select count(*) from users where is_followed = 1")" = "0"

# TODO: Maybe this file should be broken up into multiple test scripts

echo -e "\033[32mAll tests passed.  Finished successfully.\033[0m"
