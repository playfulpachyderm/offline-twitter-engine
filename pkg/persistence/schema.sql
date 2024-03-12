PRAGMA foreign_keys = on;


-- Users
-- -----

create table users (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    display_name text not null,
    handle text unique not null,
    bio text,
    following_count integer,
    followers_count integer,
    location text,
    website text,
    join_date integer,
    is_private boolean default 0,
    is_verified boolean default 0,
    is_banned boolean default 0,
    is_deleted boolean default 0,
    profile_image_url text,
    profile_image_local_path text,
    banner_image_url text,
    banner_image_local_path text,
    pinned_tweet_id integer check(typeof(pinned_tweet_id) = 'integer' or pinned_tweet_id = ''),

    is_followed boolean default 0,
    is_id_fake boolean default 0,
    is_content_downloaded boolean default 0
);

create table fake_user_sequence(latest_fake_id integer not null);
insert into fake_user_sequence values(0x4000000000000000);

create table lists(rowid integer primary key,
    is_online boolean not null default 0,
    online_list_id integer not null default 0, -- Will be 0 for lists that aren't Twitter Lists
    name text not null,
    check ((is_online = 0 and online_list_id = 0) or (is_online != 0 and online_list_id != 0))
    check (rowid != 0)
);
create table list_users(rowid integer primary key,
    list_id integer not null,
    user_id integer not null,
    unique(list_id, user_id)
    foreign key(list_id) references lists(rowid) on delete cascade
    foreign key(user_id) references users(id)
);
create index if not exists index_list_users_list_id on list_users (list_id);
create index if not exists index_list_users_user_id on list_users (user_id);
insert into lists(rowid, name) values (1, "Offline Follows");
insert into list_users(list_id, user_id) select 1, id from users where is_followed = 1;

create table follows(rowid integer primary key,
    follower_id integer not null,
    followee_id integer not null,
    unique(follower_id, followee_id),
    foreign key(follower_id) references users(id)
    foreign key(followee_id) references users(id)
);
create index if not exists index_follows_followee_id on follows (followee_id);
create index if not exists index_follows_follower_id on follows (follower_id);


-- Tweets
-- ------

create table tombstone_types (rowid integer primary key,
    short_name text not null unique,
    tombstone_text text not null unique
);
insert into tombstone_types(rowid, short_name, tombstone_text) values
    (1, 'deleted', 'This Tweet was deleted by the Tweet author'),
    (2, 'suspended', 'This Tweet is from a suspended account'),
    (3, 'hidden', 'You’re unable to view this Tweet because this account owner limits who can view their Tweets'),
    (4, 'unavailable', 'This Tweet is unavailable'),
    (5, 'violated', 'This Tweet violated the Twitter Rules'),
    (6, 'no longer exists', 'This Tweet is from an account that no longer exists'),
    (7, 'age-restricted', 'Age-restricted adult content. This content might not be appropriate for people under 18 years old. To view this media, you’ll need to log in to Twitter'),
    (8, 'newer-version-available', 'There’s a new version of this Tweet');

create table tweets (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    user_id integer not null check(typeof(user_id) = 'integer'),
    text text not null,
    is_expandable bool not null default 0,
    posted_at integer,
    num_likes integer,
    num_retweets integer,
    num_replies integer,
    num_quote_tweets integer,
    in_reply_to_id integer,
    quoted_tweet_id integer,
    mentions text,        -- comma-separated
    reply_mentions text,  -- comma-separated
    hashtags text,        -- comma-separated
    space_id text,
    tombstone_type integer default 0,
    is_stub boolean default 0,

    is_content_downloaded boolean default 0,
    is_conversation_scraped boolean default 0,
    last_scraped_at integer not null default 0,
    foreign key(user_id) references users(id)
    foreign key(space_id) references spaces(id)
);
create index if not exists index_tweets_in_reply_to_id on tweets (in_reply_to_id);
create index if not exists index_tweets_user_id        on tweets (user_id);
create index if not exists index_tweets_posted_at      on tweets (posted_at);


-- Tweet content
-- -------------

create table urls (rowid integer primary key,
    tweet_id integer not null,
    domain text,
    text text not null,
    short_text text not null default "",
    title text,
    description text,
    creator_id integer,
    site_id integer,
    thumbnail_width integer not null,
    thumbnail_height integer not null,
    thumbnail_remote_url text,
    thumbnail_local_path text,
    has_card boolean,
    has_thumbnail boolean,
    is_content_downloaded boolean default 0,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_urls_tweet_id on urls (tweet_id);

create table polls (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    tweet_id integer not null,
    num_choices integer not null,

    choice1 text,
    choice1_votes integer,
    choice2 text,
    choice2_votes integer,
    choice3 text,
    choice3_votes integer,
    choice4 text,
    choice4_votes integer,

    voting_duration integer not null,  -- in seconds
    voting_ends_at integer not null,

    last_scraped_at integer not null,

    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_polls_tweet_id on polls (tweet_id);


create table images (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    tweet_id integer not null,
    width integer not null,
    height integer not null,
    remote_url text not null unique,
    local_filename text not null unique,
    is_downloaded boolean default 0,

    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_images_tweet_id on images (tweet_id);

create table videos (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    tweet_id integer not null,
    width integer not null,
    height integer not null,
    remote_url text not null unique,
    local_filename text not null unique,
    thumbnail_remote_url text not null default "missing",
    thumbnail_local_filename text not null default "missing",
    duration integer not null default 0,
    view_count integer not null default 0,
    is_gif boolean default 0,
    is_downloaded boolean default 0,
    is_blocked_by_dmca boolean not null default 0,

    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_videos_tweet_id on videos (tweet_id);

create table hashtags (rowid integer primary key,
    tweet_id integer not null,
    text text not null,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);


-- Retweets
-- --------

create table retweets(rowid integer primary key,
    retweet_id integer not null unique,
    tweet_id integer not null,
    retweeted_by integer not null,
    retweeted_at integer not null,
    foreign key(tweet_id) references tweets(id)
    foreign key(retweeted_by) references users(id)
);
create index if not exists index_retweets_retweeted_at on retweets (retweeted_at);


-- Spaces
-- ------

create table spaces(rowid integer primary key,
    id text unique not null,
    created_by_id integer,
    short_url text not null,
    state text not null,
    title text not null,
    created_at integer not null,
    started_at integer not null,
    ended_at integer not null,
    updated_at integer not null,
    is_available_for_replay boolean not null,
    replay_watch_count integer,
    live_listeners_count integer,
    is_details_fetched boolean not null default 0,

    foreign key(created_by_id) references users(id)
);

create table space_participants(rowid integer primary key,
    user_id integer not null,
    space_id not null,

    unique(user_id, space_id)
    foreign key(space_id) references spaces(id)
    -- No foreign key for users, since they may not be downloaded yet and I don't want to
    -- download every user who joins a space
);


-- Likes
-- -----

create table likes(rowid integer primary key,
    sort_order integer not null, -- Can't be unique because "-1" is used as "unknown" value
    user_id integer not null,
    tweet_id integer not null,
    unique(user_id, tweet_id)
    foreign key(user_id) references users(id)
    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_likes_user_id on likes (user_id);
create index if not exists index_likes_tweet_id on likes (tweet_id);


-- Direct Messages (DMs)
-- ---------------------

create table chat_rooms (rowid integer primary key,
    id text unique not null,
    type text not null,
    last_messaged_at integer not null,
    is_nsfw boolean not null,

    -- Group DM info
    created_at integer not null,
    created_by_user_id integer not null,
    name text not null default '',
    avatar_image_remote_url text not null default '',
    avatar_image_local_path text not null default ''
);

create table chat_room_participants(rowid integer primary key,
    chat_room_id text not null,
    user_id integer not null,
    last_read_event_id integer not null,
    is_chat_settings_valid boolean not null default 0,
    is_notifications_disabled boolean not null,
    is_mention_notifications_disabled boolean not null,
    is_read_only boolean not null,
    is_trusted boolean not null,
    is_muted boolean not null,
    status text not null,
    unique(chat_room_id, user_id)
);

create table chat_messages (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    chat_room_id text not null,
    sender_id integer not null,
    sent_at integer not null,
    request_id text not null,
    in_reply_to_id integer,
    text text not null,
    embedded_tweet_id integer not null default 0,
    foreign key(chat_room_id) references chat_rooms(id)
    foreign key(sender_id) references users(id)
);

create table chat_message_reactions (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    message_id integer not null,
    sender_id integer not null,
    sent_at integer not null,
    emoji text not null,
    foreign key(message_id) references chat_messages(id)
    foreign key(sender_id) references users(id)
);

create table chat_message_images (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    chat_message_id integer not null,
    width integer not null,
    height integer not null,
    remote_url text not null unique,
    local_filename text not null unique,
    is_downloaded boolean default 0,

    foreign key(chat_message_id) references chat_messages(id)
);
create index if not exists index_chat_message_images_chat_message_id on chat_message_images (chat_message_id);

create table chat_message_videos (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    chat_message_id integer not null,
    width integer not null,
    height integer not null,
    remote_url text not null unique,
    local_filename text not null unique,
    thumbnail_remote_url text not null default "missing",
    thumbnail_local_filename text not null default "missing",
    duration integer not null default 0,
    view_count integer not null default 0,
    is_gif boolean default 0,
    is_downloaded boolean default 0,
    is_blocked_by_dmca boolean not null default 0,

    foreign key(chat_message_id) references chat_messages(id)
);
create index if not exists index_chat_message_videos_chat_message_id on chat_message_videos (chat_message_id);

create table chat_message_urls (rowid integer primary key,
    chat_message_id integer not null,
    domain text,
    text text not null,
    short_text text not null default "",
    title text,
    description text,
    creator_id integer,
    site_id integer,
    thumbnail_width integer not null,
    thumbnail_height integer not null,
    thumbnail_remote_url text,
    thumbnail_local_path text,
    has_card boolean,
    has_thumbnail boolean,
    is_content_downloaded boolean default 0,

    unique (chat_message_id, text)
    foreign key(chat_message_id) references chat_messages(id)
);
create index if not exists index_chat_message_urls_chat_message_id on chat_message_urls (chat_message_id);


-- Meta
-- ----

create table database_version(rowid integer primary key,
    version_number integer not null unique
);
