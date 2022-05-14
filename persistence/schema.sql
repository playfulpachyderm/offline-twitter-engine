PRAGMA foreign_keys = on;

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
    profile_image_url text,
    profile_image_local_path text,
    banner_image_url text,
    banner_image_local_path text,
    pinned_tweet_id integer check(typeof(pinned_tweet_id) = 'integer' or pinned_tweet_id = ''),

    is_followed boolean default 0,
    is_id_fake boolean default 0,
    is_content_downloaded boolean default 0
);

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
    (7, 'age-restricted', 'Age-restricted adult content. This content might not be appropriate for people under 18 years old. To view this media, you’ll need to log in to Twitter');

create table tweets (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    user_id integer not null check(typeof(user_id) = 'integer'),
    text text not null,
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
    tombstone_type integer default 0,
    is_stub boolean default 0,

    is_content_downloaded boolean default 0,
    is_conversation_scraped boolean default 0,
    last_scraped_at integer not null default 0,
    foreign key(user_id) references users(id)
);

create table retweets(rowid integer primary key,
    retweet_id integer not null unique,
    tweet_id integer not null,
    retweeted_by integer not null,
    retweeted_at integer not null,
    foreign key(tweet_id) references tweets(id)
    foreign key(retweeted_by) references users(id)
);

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

create table spaces(rowid integer primary key,
    id text unique not null,
    short_url text not null
);

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

    foreign key(tweet_id) references tweets(id)
);

create table hashtags (rowid integer primary key,
    tweet_id integer not null,
    text text not null,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);

create table database_version(rowid integer primary key,
    version_number integer not null unique
);

create table fake_user_sequence(latest_fake_id integer not null);
insert into fake_user_sequence values(0x4000000000000000);
