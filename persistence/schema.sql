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
    profile_image_url text,
    banner_image_url text,
    pinned_tweet_id integer check(typeof(pinned_tweet_id) = 'integer' or pinned_tweet_id = ''),

    is_content_downloaded boolean default 0

    -- foreign key(pinned_tweet_id) references tweets(id)
);

create table tweets (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    user_id integer not null check(typeof(user_id) = 'integer'),
    text text not null,
    posted_at integer,
    num_likes integer,
    num_retweets integer,
    num_replies integer,
    num_quote_tweets integer,
    in_reply_to integer,
    quoted_tweet integer,
    mentions text,  -- comma-separated
    hashtags text,  -- comma-separated

    is_content_downloaded boolean default 0,
    foreign key(user_id) references users(id)
    -- foreign key(in_reply_to) references tweets(id),
    -- foreign key(quoted_tweet) references tweets(id)
);

create table retweets(rowid integer primary key,
    retweet_id integer not null,
    tweet_id integer not null,
    retweeted_by integer not null,
    retweeted_at integer not null,
    foreign key(tweet_id) references tweets(id)
    foreign key(retweeted_by) references users(id)
);

create table urls (rowid integer primary key,
    tweet_id integer not null,
    text text not null,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);

create table images (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    tweet_id integer not null,
    filename text not null unique,
    is_downloaded boolean default 0,

    foreign key(tweet_id) references tweets(id)
);

create table videos (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    tweet_id integer not null,
    remote_url text not null unique,
    local_filename text not null unique,
    is_downloaded boolean default 0,

    foreign key(tweet_id) references tweets(id)
);

create table hashtags (rowid integer primary key,
    tweet_id integer not null,
    text text not null,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);
