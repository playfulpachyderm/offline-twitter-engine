.bail on
PRAGMA foreign_keys=ON;

BEGIN TRANSACTION;
CREATE TABLE users (rowid integer primary key,
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

    is_id_fake boolean default 0,
    is_followed boolean default 0,
    is_content_downloaded boolean default 0,
    is_deleted boolean default 0
);
INSERT INTO users VALUES
    (1,2703181339,'Denlesks','Denlesks','Parody News.     I was born to rock the boat, some may sink but we will float, grab your coat let’s get out of here',197,11153,'California','',1407036594000,0,0,0,'https://pbs.twimg.com/profile_images/1245461144539516930/22YJvhC7.jpg','Denlesks_profile_22YJvhC7.jpg','https://pbs.twimg.com/profile_banners/2703181339/1585776052','Denlesks_banner_1585776052.jpg',1030981649935216640,0,0,0,0),
    (2,1243523149712556032,'Metadoxy','Xirong7',replace('harbinger of the triumph of the designed over the designoid.\n Player of the glass bead game, Autopoietic Turboencabulator','\n',char(10)),4829,2123,'','',1585314008000,0,0,0,'https://pbs.twimg.com/profile_images/1342955019767869446/YSVzIrl4.jpg','Xirong7_profile_YSVzIrl4.jpg','https://pbs.twimg.com/profile_banners/1243523149712556032/1608906491','Xirong7_banner_1608906491.jpg',1352393109200580608,0,0,0,0),
    (3,1032468021485293568,'Nemets','Peter_Nimitz','Interested in wild adventures, obscure tribes, & historical processes.',268,19739,'Las Vegas, USA','',1534994533000,0,0,0,'https://pbs.twimg.com/profile_images/1036304583247552512/ro1SuGao.jpg','Peter_Nimitz_profile_ro1SuGao.jpg','https://pbs.twimg.com/profile_banners/1032468021485293568/1553221184','Peter_Nimitz_banner_1553221184.jpg',1412320347404136452,0,1,0,0),
    (4,96906231,'Viva Frei','thevivafrei','Lawyer, YouTuber, Shorty Award Winner, GoPro Award Winner,cool dad, annoying husband, drone enthusiast, GoPro maniac, eccentric. YouTube: https://t.co/oVDb0G6BmN',441,52096,'Canada','https://www.vivabarneslaw.locals.com',1260848775000,0,0,0,'https://pbs.twimg.com/profile_images/1299069199919263750/sgMtqX08.jpg','thevivafrei_profile_sgMtqX08.jpg','https://pbs.twimg.com/profile_banners/96906231/1594950253','thevivafrei_banner_1594950253.jpg',1426357774531891200,0,0,0,0),
    (5,836779281049014272,'Bronze Age Kashi','kwamurai','Comic Mishimist. Internecromancer. ~mirtyd-pasleg',370,11704,'','',1488338702000,0,0,0,'https://pbs.twimg.com/profile_images/1424568508747223044/3qS9O7Np.jpg','kwamurai_profile_3qS9O7Np.jpg','https://pbs.twimg.com/profile_banners/836779281049014272/1611435371','kwamurai_banner_1611435371.jpg',1423000763358404610,0,0,1,0),
    (6,1109842387365433344,'Cordón de Yute','Merde22898677',replace('Keep clear of the dupes that talk democracy\nAnd the dogs that talk revolution,\nDrunk with talk, liars and believers.\nI believe in my tusks. -Robinson Jeffers','\n',char(10)),433,148,'','https://cord-of-jute.blogspot.com/?m=1',1553442019000,0,0,0,'https://pbs.twimg.com/profile_images/1388278226657611780/Wk376nt5.jpg','Merde22898677_profile_Wk376nt5.jpg','https://pbs.twimg.com/profile_banners/1109842387365433344/1619826432','Merde22898677_banner_1619826432.jpg',1299884979321581569,0,0,0,0),
    (7,887434912529338375,'Bronze Age Covfefe Anon','CovfefeAnon',replace('Not to be confused with 2001 Nobel Peace Prize winner Kofi Annan.\n\n54th Clause of the Magna Carta absolutist.\n\nCommentary from an NRx perspective.','\n',char(10)),469,5848,'','',1500415945000,0,0,0,'https://pbs.twimg.com/profile_images/1392509603116617731/TDrNeUiZ.jpg','CovfefeAnon_profile_TDrNeUiZ.jpg','https://pbs.twimg.com/profile_banners/887434912529338375/1598514714','CovfefeAnon_banner_1598514714.jpg',1005906691324596224,0,0,0,0),
    (8,1282037650384461825,'Charles','JiffjoffI',replace('Statistician working in BigTech; \nPosting on History, film, baseball, culture, dogs','\n',char(10)),463,246,'Clown World, USA','',1594496577000,0,0,0,'https://pbs.twimg.com/profile_images/1345679237865463809/qeZHMIjC.jpg','JiffjoffI_profile_qeZHMIjC.jpg','','',1307777709024645124,0,0,0,0),
    (9,1067869346775646208,'Shashank Nayak','ShazCoder','Software Engineer. Interested in Economic History, Finance and Programming.',194,679,'Mumbai, India','',1543434867000,0,0,0,'https://pbs.twimg.com/profile_images/1097620512635797504/VdSzR2Op.png','ShazCoder_profile_VdSzR2Op.png','','',0,0,0,0,0),
    (10,1372116552942764034,'Indo-Germanic','Germany12343','interbol agent',192,58,'','',1615973080000,0,0,0,'https://pbs.twimg.com/profile_images/1372219154237943814/Fo4dhnPw.jpg','Germany12343_profile_Fo4dhnPw.jpg','https://pbs.twimg.com/profile_banners/1372116552942764034/1615997697','Germany12343_banner_1615997697.jpg',1421965055508328450,0,0,0,0),
    (11,1304281147074064385,'Urban Artisan','artisan_urban','The status quo and episteme must be radically changed for the health of the body and soul.',825,228,'Empire of the Setting Sun','',1599799844000,0,0,0,'https://pbs.twimg.com/profile_images/1317983221062258691/aV__4fXd.jpg','artisan_urban_profile_aV__4fXd.jpg','https://pbs.twimg.com/profile_banners/1304281147074064385/1599804430','artisan_urban_banner_1599804430.jpg',1317985338288181248,0,1,0,0),
    (1093,1357717304931479552,'𝘪𝘯𝘥𝘪𝘢𝘯 𝘣𝘳𝘰𝘯𝘴𝘰𝘯','lndian_Bronson',replace('https://t.co/myFkyKG1KJ\n\nhttps://t.co/SN8lIlBeQu','\n',char(10)),2820,8321,'','',1612540031000,0,0,0,'https://pbs.twimg.com/profile_images/1439862664197443585/Tb6Q9A7g.jpg','lndian_Bronson_profile_Tb6Q9A7g.jpg','https://pbs.twimg.com/profile_banners/1357717304931479552/1631923651','lndian_Bronson_banner_1631923651.jpg',1365487261454901248,0,0,0,0),
    (16,358545917,'Cernovich','Cernovich','Filmmaker, watch my latest documentary on fake news, Hoaxed. Clink link below.',3066,763726,'Orange County, California','https://itunes.apple.com/us/movie/hoaxed/id1498889845',1313809349000,0,1,0,'https://pbs.twimg.com/profile_images/1431989112874024964/SzaC_Qmk.jpg','Cernovich_profile_SzaC_Qmk.jpg','https://pbs.twimg.com/profile_banners/358545917/1628836463','Cernovich_banner_1628836463.jpg',1439747634277740546,0,1,0,0),
    (1064,4731154187,'Sam Considine','s_considine1','Anti-Lockdown Crusader Fighting To Restore Our Basic Liberties. Views Are My Own, Why Give Someone Else Credit For Being Right?',833,1071,'New York, USA','',1452140589000,0,0,0,'https://pbs.twimg.com/profile_images/1387628943260459008/yI2X3lSr.jpg','s_considine1_profile_yI2X3lSr.jpg','https://pbs.twimg.com/profile_banners/4731154187/1620032248','s_considine1_banner_1620032248.jpg',1204371476549455872,0,0,0,0),
    (2001,44067298,'Michael Malice','michaelmalice',replace('Author of Dear Reader, The New Right & The Anarchist Handbook\nHost of "YOUR WELCOME" \nSubject of Ego & Hubris by Harvey Pekar\nHe/Him ⚑\n@SheathUnderwear Model','\n',char(10)),968,270826,'Austin','https://amzn.to/3oInafv',1243920952000,0,1,0,'https://pbs.twimg.com/profile_images/1415820415314931715/_VVX4GI8.jpg','michaelmalice_profile__VVX4GI8.jpg','https://pbs.twimg.com/profile_banners/44067298/1615134676','michaelmalice_banner_1615134676.jpg',1426357554741981184,0,0,0,0),
    (1145,14347972,'Mark Hemingway','Heminator','Senior Writer at RealClearInvestigations. "After all these years of professional experience, why can’t I write good?" Husband of @MZHemingway.',5544,86321,'','http://www.weeklystandard.com/rough-draft/article/2010315',1207796678000,0,1,0,'https://pbs.twimg.com/profile_images/555863013027094528/MUOYcD2g.png','Heminator_profile_MUOYcD2g.png','https://pbs.twimg.com/profile_banners/14347972/1532479949','Heminator_banner_1532479949.jpg',0,1,1,0,0),
    (175465,1427250806378672134,'','nancytracker','',0,0,'','',-62135596800,0,0,1,'','','','',0,1,0,0,0),
    (175466,2579616110,'iamhammed','iamhammed_','',296,161,'','',1403315832000,0,0,0,'https://pbs.twimg.com/profile_images/1467990006929268743/flZXQYm1.jpg','iamhammed__profile_flZXQYm1.jpg','','',0,0,0,0,0),
    (175520,18812728,'Andrew Schulz 👑HEZI','andrewschulz','Comedian. Podcasts: Flagrant 2 & The Brilliant Idiots IG: AndrewSchulz Bookings@TheAndrewSchulz.com',757,308546,'New York, NY','http://youtube.com/TheAndrewSchulz',1231530798000,0,1,0,'https://pbs.twimg.com/profile_images/1082514379176923136/dzlt77WJ.jpg','andrewschulz_profile_dzlt77WJ.jpg','https://pbs.twimg.com/profile_banners/18812728/1608052737','andrewschulz_banner_1608052737.jpg',1394326801510551553,0,0,0,0),
    (3180,1178839081222115328,'Mystery Grove Publishing Co.','MysteryGrove',replace('Featured books:\nThe Storm of Steel: https://t.co/UH7zDOI8Dh\nAlways with Honor: https://t.co/zNDbP5Xz3n\nMine Were of Trouble: https://t.co/MqVgqZOUuB\n\nFull catalog: https://t.co/o3q88bFqjd','\n',char(10)),7812,25834,'','',1569892125000,0,0,0,'https://pbs.twimg.com/profile_images/1254314471813758976/sRWOQDLz.jpg','MysteryGrove_profile_sRWOQDLz.jpg','https://pbs.twimg.com/profile_banners/1178839081222115328/1592880438','MysteryGrove_banner_1592880438.jpg',1505239085778186243,1,0,0,0),
    (7041,1240784920831762433,'Lukas (computer)','SCHIZO_FREQ','Retired Engagement Agriculturalist',813,51341,'The Obelisk','https://lukasxp.substack.com',1584661213000,0,0,0,'https://pbs.twimg.com/profile_images/1603480681065103362/0BGtxtfu.jpg','SCHIZO_FREQ_profile_0BGtxtfu.jpg','https://pbs.twimg.com/profile_banners/1240784920831762433/1665972431','SCHIZO_FREQ_banner_1665972431.jpg',1524509932099448833,1,0,0,0),
    (175547,19370504,'Alexander Cortes PhD, Fitness, Nutrition, Fat loss','AJA_Cortes','#1 OG of Fitness Twitter. Doctorate in BroScience. 12 years producing physical  transformations. Build muscle, burn fat, Join 42K people reading my newsletter',1249,184562,'Weekly Newsletter','https://cortes.site/newsletter/',1232668248000,0,0,0,'https://pbs.twimg.com/profile_images/1611029834842374144/sa9CI9EP.jpg','AJA_Cortes_profile_sa9CI9EP.jpg','https://pbs.twimg.com/profile_banners/19370504/1630521966','AJA_Cortes_banner_1630521966.jpg',1695255226108813688,0,0,0,0),
    (97706,1159179478582603776,'Evelyn Kokemoor','EKokemoor','mars/wisconsin ⚧️⚢ ~macrep-racdec',256,139,'','',1565204898000,0,0,0,'https://pbs.twimg.com/profile_images/1643762712868970497/r9JyQjKg.jpg','EKokemoor_profile_r9JyQjKg.jpg','https://pbs.twimg.com/profile_banners/1159179478582603776/1626219975','EKokemoor_banner_1626219975.jpg',1540465706139090944,0,0,0,0),
    (160242,534463724,'iko','ilyakooo0',replace('Code poet.\n~racfer-hattes','\n',char(10)),473,173,'','http://iko.soy',1332519666000,0,0,0,'https://pbs.twimg.com/profile_images/1671427114438909952/8v8raTeb.jpg','ilyakooo0_profile_8v8raTeb.jpg','','',0,0,0,0,0),
    (169994,1689006330235760640,'sol🏴‍☠️','sol_plunder','',165,134,'','',1691525490000,0,0,0,'https://pbs.twimg.com/profile_images/1689006644905033728/T1uO4Jvt.jpg','sol_plunder_profile_T1uO4Jvt.jpg','','',1704554384930058537,0,0,0,0),
    (1680,1458284524761075714,'wispem-wantex','wispem_wantex',replace('~wispem-wantex\n\nCurrently looking for work (DMs open)','\n',char(10)),136,483,'on my computer','https://offline-twitter.com/',1636517116000,0,0,0,'https://pbs.twimg.com/profile_images/1462880679687954433/dXJN4Bo4.jpg','wispem_wantex_profile_dXJN4Bo4.jpg','','',1695221528617468324,1,0,0,0),
    (27398,1488963321701171204,'Offline Twatter','Offline_Twatter',replace('Offline Twitter is an open source twitter client and tweet-archiving app all in one.  Try it out!\n\nSource code: https://t.co/2PMumKSxFO','\n',char(10)),4,2,'','https://offline-twitter.com',1643831522000,0,0,0,'https://pbs.twimg.com/profile_images/1507883049853210626/TytFbk_3.jpg','Offline_Twatter_profile_TytFbk_3.jpg','','',1507883724615999488,1,1,0,0),
    (175560,249206992,'ludwig','ludwigABAP','God’s chosen principal engineer. What is impossible for you is not impossible for me.',984,17966,'','https://ludwigabap.bearblog.dev/',1297180819000,0,0,0,'https://pbs.twimg.com/profile_images/1753215006697459712/n76_qnTj.jpg','ludwigABAP_profile_n76_qnTj.jpg','https://pbs.twimg.com/profile_banners/249206992/1706835247','ludwigABAP_banner_1706835247.jpg',0,0,0,0,0);


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
insert into lists(rowid, name) values (2, "Bronze Age");
insert into list_users(list_id, user_id) select 2, id from users where display_name like "%bronze age%";
insert into lists(rowid, name) values (3, "Covfefe");
insert into list_users(list_id, user_id) values (3, 887434912529338375);


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
INSERT INTO spaces VALUES
    (323,'1OwGWwnoleRGQ',1178839081222115328,'https://t.co/kxr7O7hfJ6','Ended','I''m showering and the hot water ran out',1676225386000,1676225389000,1676235389000,1676229669000,1,11,255,1);


CREATE TABLE tweets (rowid integer primary key,
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
    space_id text,
    tombstone_type integer default 0,
    is_stub boolean default 0,

    is_content_downloaded boolean default 0,
    is_conversation_scraped boolean default 0,
    last_scraped_at integer not null default 0,
    is_expandable bool not null default 0,
    foreign key(user_id) references users(id)
    foreign key(space_id) references spaces(id)
);
create index if not exists index_tweets_in_reply_to_id on tweets (in_reply_to_id);
create index if not exists index_tweets_user_id on tweets (user_id);
create index if not exists index_tweets_posted_at on tweets (posted_at);

INSERT INTO tweets VALUES
    (1,1261483383483293700,2703181339,'These are public health officials who are making decisions about your lifestyle because they know more about health, fitness and well-being than you do',1589596050000,245,87,42,21,0,0,'','','',NULL,0,0,1,0,0,0),
    (2,1413664406995566593,1032468021485293568,'Most important lesson about government imo is that a politician or movement that wants stuff done needs to get their own guys &amp; gals jobs as bureaucrats, contractors, or consultants in appropriate government organization. If you don’t, career bureaucrats will ignore you.',1625878833000,440,68,9,5,0,1413646595493568516,'','','',NULL,0,0,0,0,0,0),
    (3,1413665734866186243,1243523149712556032,'',1625879150000,2,0,0,0,1413664406995566593,0,'Peter_Nimitz','Peter_Nimitz','',NULL,0,0,0,0,0,0),
    (4,1413646595493568516,1032468021485293568,'Learned a lot about how government actually works too. Or how in California Department of Transportation’s case, doesn’t work at all.',1625874587000,184,4,4,1,1413646309047767042,0,'','','',NULL,0,0,0,1,1629035456000,0),
    (5,1426619468327882761,96906231,'The streets of Montreal today',1628967561000,6231,1640,152,98,0,0,'','','',NULL,0,0,0,0,0,0),
    (6,1343715029707796489,1109842387365433344,'"We have come to recognize that the political is the total, and as a result we know that any decision about whether something is unpolitical is always a political decision, irrespective of who decides and what reasons are advanced."  Carl Schmitt.',1609201602000,2,0,0,0,1343626462868484102,0,'kwamurai,Saradin1337','kwamurai,Saradin1337','',NULL,0,0,0,0,0,0),
    (7,1343633011364016128,836779281049014272,'this is why the "think tank mindset" is a dead end. it misapprehends the nature of power. the "battle of ideas" is a meaningless sideshow when the terms on which it is fought are set elsewhere. it is a fiction. appealing because of its simplicity but always won or lost in advance',1609182048000,138,9,2,1,1343630971057418240,0,'','','',NULL,0,0,0,0,0,0),
    (8,1426669666928414720,887434912529338375,replace('The system already gives free healthcare and college to its clients.\n\nWho could the system tax to pay for free healthcare and college for whites?','\n',char(10)),1628979529000,147,17,3,0,0,1426654719183835136,'','','',NULL,0,0,0,0,0,0),
    (2519,1428939163961790466,1282037650384461825,replace('At this point what can we expect I guess\n\nBut the reason this seems weird is b/c in other contexts tech companies have to jump through hoops to prove there weren''t any qualified Americans available to hire for the job to hire H1b i think - what''s the difference here then?','\n',char(10)),1629520619000,3,0,1,0,1428938327886073856,0,'CovfefeAnon,primalpoly,jmasseypoet,SpaceX','JiffjoffI,CovfefeAnon,primalpoly,jmasseypoet,SpaceX','',NULL,0,0,0,0,0,0),
    (9,1428951883058753537,887434912529338375,'Space X was an embarrassment in a lot of ways - it showed up NASA very badly.',1629523652000,4,0,0,0,1428939163961790466,0,'JiffjoffI,primalpoly,jmasseypoet,SpaceX','JiffjoffI,primalpoly,jmasseypoet,SpaceX','',NULL,0,0,0,0,0,0),
    (10,1413647919215906817,1032468021485293568,'I’ve lived here almost seven years now - met a lot of interesting people, went on some adventures, &amp; learned quite a bit I never expected to.',1625874902000,109,0,3,0,1413646595493568516,0,'','','',NULL,0,0,0,0,0,0),
    (11,1413657324267311104,1067869346775646208,'Did if affect your political views?',1625877145000,6,0,1,0,1413646595493568516,0,'Peter_Nimitz','Peter_Nimitz','',NULL,0,0,0,0,0,0),
    (12,1413658466795737091,1032468021485293568,'Yes - moderated them considerably. Harder to hate politicians once you realize they are often just spin men for totally unaccountable bureaucrats with their own interests.',1625877417000,74,4,2,0,1413657324267311104,0,'ShazCoder','ShazCoder','',NULL,0,0,0,0,0,0),
    (13,1413772782358433792,1372116552942764034,'Idk if this is relevant to your department, but what do you think about the high speed train efforts in California?',1625904672000,1,0,1,0,1413646595493568516,0,'Peter_Nimitz','Peter_Nimitz','',NULL,0,0,0,1,1629035457000,0),
    (14,1413773185296650241,1032468021485293568,'Good idea in theory, but in practice mostly graft',1625904768000,8,0,0,0,1413772782358433792,0,'Germany12343','Germany12343','',NULL,0,0,0,1,1629035458000,0),
    (15,1413650853081276421,1304281147074064385,'Would love to hear about it!',1625875602000,2,0,0,0,1413646595493568516,0,'Peter_Nimitz','Peter_Nimitz','',NULL,0,0,0,0,0,0),
    (2761,1413646309047767042,1032468021485293568,'Last 15 minutes of work. Pretty fortunate to have gotten a job here - liked all of my coworkers &amp; bosses even if we had our disagreements.',1625874519000,203,4,7,0,0,0,'','','',NULL,0,0,0,0,0,0),
    (147,1438642143170646017,1357717304931479552,replace('https://t.co/X1YFCSYlKh\n\nhttps://t.co/dNTDGYkJ9y\n\nhttps://t.co/Ti54Xr68dK\n\nBiden won those voters, complete with ''in this house we believe in science'' lawn posters','\n',char(10)),1631833990000,46,0,3,0,1438640730281496576,0,'ScottMGreer','ScottMGreer','',NULL,0,0,0,0,0,0),
    (2673,1439027915404939265,358545917,replace('Morally nuanced and complicated discussion.\n\nWhat do you think?','\n',char(10)),1631925965000,854,133,399,33,0,0,'','','',NULL,0,0,0,0,0,0),
    (2702,1439067163508150272,358545917,replace('I don’t think the vaccine is that risky and a lot of y’all embarrass yourselves on this. \n\nFor me the moral issue is cooperation with evil. \n\nThe vax passport is designed to exclude the “lesser” class of people. \n\nAnd where this leads to. \n\nComplicated subject.','\n',char(10)),1631935323000,413,60,169,11,0,1439027915404939265,'','','',NULL,0,0,0,0,0,0),
    (2671,1439068429768605696,4731154187,'Exactly, I actually made a vaccine appointment but canceled after visiting Florida and understanding how much freedom I already lost with enough distrust of our “experts” to know it probably wouldn’t end with a vaccine.',1631935624000,93,19,7,1,0,1439067163508150272,'','','',NULL,0,0,0,0,0,0),
    (2684,1439068749336748043,358545917,'We all draw lines. I’m fine with the vaccine. Won’t do passports or ever show proof of vaccination. That’s collaborating with evil as it’s denying services to a “lesser class.”',1631935701000,598,96,38,6,0,1439068429768605696,'','','',NULL,0,0,0,0,0,0),
    (2927,1449148515918270475,14347972,'LOL',1634338904000,81194,13586,632,608,0,0,'','','',NULL,0,0,1,0,0,0),
    (3030,1453461248142495744,358545917,'',1635367140000,85,8,7,0,0,1453452754547060736,'','','',NULL,0,0,1,0,0,0),
    (202,1465534109573390348,44067298,'Which of these tattoos would you get if you had to get one on your arm?',1638245534000,116,13,1,17,0,0,'','','',NULL,0,0,1,1,1640394060000,0),
    (2857234,31,14347972,"",1634338900000,23,24,25,26,0,0,'','','',NULL,1,1,0,0,0,0),  -- This isn't a real tweet
    (2857235,1413666994876936198,2579616110,'Good insight.',1625879450000,4,0,0,0,1413658466795737091,0,'Peter_Nimitz,ShazCoder','Peter_Nimitz,ShazCoder','',NULL,0,0,1,1,1642640600000,0),
    (2857390,1490120332484972549,18812728,'“In the end it’s not the words of our enemies we will remember but the silence of our friends.”',1644107347000,5798,770,106,37,0,0,'','','',NULL,NULL,0,1,0,0,0),
    (2857409,1490116725395927042,18812728,replace('Rogan has made a lot of people millionaires. Imagine being one of those people and staying silent right now? \n\nCause this will blow over in a month but that silence will never be forgotten.','\n',char(10)),1644106487000,12264,1387,273,80,0,0,'','','',NULL,NULL,0,1,0,0,0),
    (2857357,1489944024278523906,96906231,'According to @gofundme it was "as a result of multiple discussions with locals law enforcement and *police reports of violence and other unlawful activity*". ABSOLUTE LIES! I asked police officers live  and they CONFIRMED there was no violence. Pure censorship. #BankruptGoFundMe',1644065311000,5753,2127,219,110,0,0,'gofundme','','BankruptGoFundMe',NULL,NULL,0,1,0,0,0),
    (121936,1513313535480287235,1178839081222115328,'Smh wish I could RT',1649637037000,4,0,1,0,1513312559981551619,0,'PublicAnthony','PublicAnthony','',NULL,NULL,0,1,0,0,0),
    (869468,1624833173514293249,1240784920831762433,'',1676225391000,1,0,0,0,0,0,'','','','1OwGWwnoleRGQ',NULL,0,1,0,0,0),
    (2090918,1439747634277740546,358545917,'Explain why staff but not talent at these events have to wear masks. Using science.',1632097559000,29832,4145,783,163,0,0,'','','',NULL,NULL,0,1,1,1710818767264,0),
    (2857431,1695110851324256692,19370504,replace('My dad was a doctor, he retired this past year \n\nHe’s been healthy his whole life, and he saw the titanic shift (no pun intended) in obesity being normalized in real time \n\nIt used to be a 300lb person was uncommon \n\nThen it was 400lbs\n\nThen 500lbs\n\nHospitals had to upgrade their scales to veterinary scales they use in zoos,\nThat’s how fat people became \n\nObese patients would be OFFENDED if you suggested they lose weight \n\nThey would complain if you told them their back pain was because their BMI was 45 \n\nThey’d ignore all suggestions of exercise or diet and complain why can’t they just take a pill \n\nThis wasn’t outliers, this is at least 50% of the population \n\nUntil you work with general public, you cannot fully conceive the existent of people’s sloth and apathy towards their own quality of life','\n',char(10)),1692980895000,1894,224,137,25,0,0,'','','',NULL,NULL,0,1,1,1693055764000,1),
    (1405789,1698426460061487546,1458284524761075714,'Zig''s "comptime" leads to the most elegant reflection code I''ve ever seen.  It''s much cleaner and more expressive than, e.g., Python''s various __methods__, or worse, the deranged "metaclasses" nonsense; but it also has no runtime cost!',1693771397000,6,0,1,1,0,1692962678824648811,'','','',NULL,NULL,0,1,0,0,0),
    (1408662,1698762403163304110,1458284524761075714,replace('Another very cool use of Zig''s "comptime" is it lets you write real, compiled mini-languages in strings; e.g.:\n\n- SQL prepared statements\n- "printf" style format strings\n- regexps\n\nEvery language uses these, but they''re interpreted at runtime, even in compiled languages.','\n',char(10)),1693851493000,7,2,3,0,0,1698426460061487546,'','','',NULL,NULL,0,1,0,0,0),
    (1408663,1698762405268902217,1458284524761075714,'These types of operations are actually their own little programs with their own grammars.  The status quo is to embed them as string literals-- effectively, source code-- in another program, because most languages don''t have a way to handle this kind of thing cleanly.',1693851493000,0,0,1,0,1698762403163304110,0,'','','',NULL,NULL,0,1,0,0,0),
    (1408664,1698762406929781161,1458284524761075714,replace('Then the "outer program" has to essentially include a compiler for the mini-language, and at runtime it compiles and runs the mini-program.\n\nBut there''s benefits to actually compiling stuff at compile time!\n- syntax checking\n- type checking\n- runtime performance','\n',char(10)),1693851493000,0,0,1,0,1698762405268902217,0,'','','',NULL,NULL,0,1,0,0,0),
    (1408665,1698762408410390772,1458284524761075714,'There''s some lame attempts to do this in limited contexts.  Some languages (e.g., Go) check printf strings at compile time.  Or you can make linters to do static analysis for these really tiny mini-languages where parsing them is trivial (e.g., printf or regexp).',1693851494000,0,0,1,0,1698762406929781161,0,'','','',NULL,NULL,0,1,0,0,0),
    (1408666,1698762409974857832,1458284524761075714,replace('But I''m not aware of any language that can statically check SQL prepared statements, for example; or something more complicated, like an HTML templating engine.\n\nWith Zig "comptime", you could do this.','\n',char(10)),1693851494000,0,0,2,0,1698762408410390772,0,'','','',NULL,NULL,0,1,0,0,0),
    (1408667,1698762411853971851,1458284524761075714,replace('In fact you could write a compiler for any mini-language you want, include it in a Zig program, and then use that mini-language via strings in Zig code and the Zig compiler will compile it for you.\n\nIn the extreme, you could probably do some r*tarded things with this.','\n',char(10)),1693851495000,0,0,1,0,1698762409974857832,0,'','','',NULL,NULL,0,1,0,0,0),
    (1408659,1698762413393236329,1458284524761075714,replace('For example, if you wrote a compiler for another programming language-- e.g., Python-- in Zig, you could embed entire Python programs as strings and compile them into a standalone executable binary.\n\nMore interestingly, you could call functions back and forth between the two.','\n',char(10)),1693851495000,2,0,1,0,1698762411853971851,0,'','','',NULL,NULL,0,1,1,1693851886000,0),
    (1408657,1698762414957666416,1458284524761075714,'There''s probably something even dumber you could do here using Large Language Models, if you''re creative (and dumb) enough.',1693851495000,3,0,0,0,1698762413393236329,0,'','','',NULL,NULL,0,1,0,0,0),
    (2857450,1698762414957666417,1458284524761075714,'Fake tweet that should not be part of the "thread"',1693851496000,0,0,0,0,1698762405268902217,0,'','','',NULL,NULL,0,1,0,0,0),
    (1409531,1698792233619562866,534463724,replace('https://t.co/KU3C7bcqR7\n\nSame thing but 20 years earlier. \n\nAnd it''s actually used in production code.','\n',char(10)),1693858605000,3,0,1,0,1698762403163304110,0,'wispem_wantex','wispem_wantex','',NULL,NULL,0,1,0,0,0),
    (1408668,1698764077458202845,1159179478582603776,'Can you believe people actually used to use m4 and C preprocessor for this stuff? Hell.',1693851892000,1,0,1,0,1698762409974857832,0,'wispem_wantex','wispem_wantex','',NULL,NULL,0,1,1,1693852276000,0),
    (1408673,1698765208393576891,1458284524761075714,'I appreciate the C preprocessor for this cutting insight',1693852161000,0,0,0,0,1698764077458202845,1620206875450167296,'EKokemoor','EKokemoor','',NULL,NULL,0,1,0,0,0),
    (1409940,1698797388914151523,1458284524761075714,replace('This looks quite neat, but "comptime" is cool because it was designed to do stuff like declaring arrays where the size is the result of a function call, e.g.\n\nvar my_array: [fibonacci(10)]u32;\n\n...yet being able to create DSLs just emerged from this very simple concept','\n',char(10)),1693859834000,2,0,1,0,1698792233619562866,0,'ilyakooo0','ilyakooo0','',NULL,NULL,0,1,0,0,0),
    (1409953,1698802806096846909,1689006330235760640,replace('Just poking around at some examples and explanation videos, It does seem very similar to Template Haskell, though maybe a bit more ergonomic.\n\nIs there something missing from this mental model?','\n',char(10)),1693861125000,3,0,1,0,1698797388914151523,0,'wispem_wantex,ilyakooo0','wispem_wantex,ilyakooo0','',NULL,NULL,0,1,0,0,0),
    (1411566,1698848086880133147,1458284524761075714,'I have basically no experience with one and literally no experience with the other, and additionally I''ve never even used Haskell.  So unfortunately I''m not really in a position to say.',1693871921000,1,0,1,0,1698802806096846909,0,'sol_plunder,ilyakooo0','sol_plunder,ilyakooo0','',NULL,NULL, 0,1,0,0,0),
    (1169437,1665509126737129472,1458284524761075714,replace('Btw, to the extent this has happened, it''s partly thanks to the Golden One (@TheGloriousLion) who invented #fizeekfriday and the "post physique" rejoinder.  Everyone should follow him if they don''t already.\n\nSince I forgot last week, and since it''s topical, here''s a leg poast','\n',char(10)),1685923294000,7,0,0,0,1665505986184900611,0,'TheGloriousLion','','fizeekfriday',NULL,NULL,0,1,0,0,0),
    (2857438,1826778617705115868,1488963321701171204,'Conversations are trees, not sequences.  They branch.  They don''t flow in a perfectly linear way.',1724372937000,4,1,0,0,0,0,'','','',NULL,NULL,0,1,0,0,0),
    (2857439,1826778617705115869,1178839081222115328,'Real tweet that is definitely real',1724372938000,4,1,0,0,1826778617705115868,0,'Offline_Twatter','Offline_Twatter','',NULL,NULL,0,1,0,0,0),
    (98105,1507883724615999488,1488963321701171204,'Apparently I am not allowed to call myself "Twitter"',1648342469000,5,1,0,1,0,0,'','','',NULL,NULL,0,1,1,1739654377321,0);

CREATE TABLE retweets(rowid integer primary key,
    retweet_id integer not null unique,
    tweet_id integer not null,
    retweeted_by integer not null,
    retweeted_at integer not null,

    foreign key(tweet_id) references tweets(id)
    foreign key(retweeted_by) references users(id)
);
create index if not exists index_retweets_retweeted_at on retweets (retweeted_at);

INSERT INTO retweets VALUES
    (33,144919526660333333,1465534109573390348,1304281147074064385,1625877777000), -- This is fake
    (15,1449195266603630594,1449148515918270475,44067298,1634350050000),
    (52,1490135787144237058,1490120332484972549,358545917,1644111031000),
    (42,1490119308692766723,1490116725395927042,358545917,1644107102000),
    (59,1490100255987171332,1489944024278523906,358545917,1644102560000),
    (1000,1490135787124232223,1698762413393236329,1488963321701171204,1644111022000), -- This is fake
    (1002,1490135787124232225,1413664406995566593,1178839081222115328,1644111025000); -- This is fake

create table urls (rowid integer primary key,
    tweet_id integer not null,
    domain text,
    text text not null,
    short_text text not null default "",
    title text,
    description text,
    creator_id integer,
    site_id integer,
    thumbnail_width integer,
    thumbnail_height integer,
    thumbnail_remote_url text,
    thumbnail_local_path text,
    has_card boolean,
    has_thumbnail boolean,
    is_content_downloaded boolean default 0,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_urls_tweet_id on urls (tweet_id);

INSERT INTO urls VALUES
    (1,1413665734866186243,'en.m.wikipedia.org','https://en.m.wikipedia.org/wiki/Entryism','','Entryism - Wikipedia','',0,0,0,0,'','',1,0,0),
    (23,1438642143170646017,'','https://www.politico.com/story/2016/07/joe-biden-democrats-middle-class-226306','','','',0,0,0,0,'','',0,0,0),
    (24,1438642143170646017,'','https://time.com/5878437/trump-white-middle-class-voters/','','','',0,0,0,0,'','',0,0,0),
    (25,1438642143170646017,'www.brookings.edu','https://www.brookings.edu/research/bidens-victory-came-from-the-suburbs/','','Biden’s victory came from the suburbs','New data reveal that Trump’s loss to Joe Biden was due mostly to voters in large metropolitan suburbs, especially in important battleground states, William Frey analyzes.',0,151106990,568,320,'https://pbs.twimg.com/card_img/1439394661521625090/W2kzjt4-?format=jpg&name=800x320_1','W2kzjt4-_800x320_1.jpg',1,1,0);


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
INSERT INTO polls VALUES
    (3,1465534108923314180,1465534109573390348,4,'Tribal armband',1593,'Marijuana leaf',624,'Butterfly',778,'Maple leaf',1138,86400,1638331934000,1638331935000);


create table space_participants(rowid integer primary key,
    user_id integer not null,
    space_id not null,

    foreign key(space_id) references spaces(id)
    -- No foreign key for users, since they may not be downloaded yet and I don't want to
    -- download every user who joins a space
);

INSERT INTO space_participants VALUES
    (411027,238001308,'1OwGWwnoleRGQ'),
    (411135,555353627,'1OwGWwnoleRGQ'),
    (410975,1012772213471105024,'1OwGWwnoleRGQ'),
    (411028,1233808749887922178,'1OwGWwnoleRGQ'),
    (410974,1240784920831762433,'1OwGWwnoleRGQ'),
    (411306,1489176151046213633,'1OwGWwnoleRGQ'),
    (411192,1620533013565083648,'1OwGWwnoleRGQ'),
    (411029,1622390441458151424,'1OwGWwnoleRGQ'),
    (411190,1623438835295342592,'1OwGWwnoleRGQ');


CREATE TABLE images (rowid integer primary key,
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

INSERT INTO images VALUES
    (1,1261483377363791872,1261483383483293700,1914,1456,'https://pbs.twimg.com/media/EYGwcrXUMAAiyCf.jpg','EYGwcrXUMAAiyCf.jpg',1),
    (2,1261483377368039424,1261483383483293700,1440,960,'https://pbs.twimg.com/media/EYGwcrYVAAAFY_U.jpg','EYGwcrYVAAAFY_U.jpg',1),
    (3,1261483377409970177,1261483383483293700,620,410,'https://pbs.twimg.com/media/EYGwcriU0AEvGA1.jpg','EYGwcriU0AEvGA1.jpg',1),
    (4,1261483377519017984,1261483383483293700,1200,893,'https://pbs.twimg.com/media/EYGwcr8UwAApzgz.jpg','EYGwcr8UwAApzgz.jpg',1),
    (5,1426669635450163204,1426669666928414720,0,0,'https://pbs.twimg.com/media/E8yMeYDX0AQcSAj.jpg','E8yMeYDX0AQcSAj.jpg',0);

CREATE TABLE videos (rowid integer primary key,
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

INSERT INTO videos VALUES
    (1,1426619366829924358,1426619468327882761,1280,720,'https://video.twimg.com/ext_tw_video/1426619366829924358/pu/vid/1280x720/vjY7yiXiRMV4m9T1.mp4?tag=12','1426619468327882761.mp4', 'https://pbs.twimg.com/ext_tw_video_thumb/1426619366829924358/pu/img/uGKC9nivwo1GUELy.jpg','uGKC9nivwo1GUELy.jpg',22180,185404,0,0,0),
    (20,1453461242698350592,1453461248142495744,224,126,'https://video.twimg.com/tweet_video/FCu7TKnVQAABftH.mp4','1453461248142495744.mp4','https://pbs.twimg.com/tweet_video_thumb/FCu7TKnVQAABftH.jpg','FCu7TKnVQAABftH.jpg',0,0,1,1,0);

CREATE TABLE hashtags (rowid integer primary key,
    tweet_id integer not null,
    text text not null,

    unique (tweet_id, text)
    foreign key(tweet_id) references tweets(id)
);

create table likes(rowid integer primary key,
    sort_order integer not null,
    user_id integer not null,
    tweet_id integer not null,
    unique(user_id, tweet_id)
    foreign key(user_id) references users(id)
    foreign key(tweet_id) references tweets(id)
);
create index if not exists index_likes_user_id on likes (user_id);
create index if not exists index_likes_tweet_id on likes (tweet_id);

insert into likes values
    (1, 1, 1178839081222115328, 1413646595493568516),
    (2, 2, 1178839081222115328, 1513313535480287235),
    (3, 3, 1178839081222115328, 1343633011364016128),
    (4, 4, 1178839081222115328, 1426669666928414720),
    (5, 5, 1178839081222115328, 1698765208393576891),
    (6, 6, 1488963321701171204, 1826778617705115869);


create table bookmarks(rowid integer primary key,
    sort_order integer not null,
    user_id integer not null,
    tweet_id integer not null,
    unique(user_id, tweet_id)
    foreign key(tweet_id) references tweets(id)
    foreign key(user_id) references users(id)
);
create index if not exists index_bookmarks_user_id on bookmarks (user_id);
create index if not exists index_bookmarks_tweet_id on bookmarks (tweet_id);
insert into bookmarks values
    (23,1800452344077464795,1488963321701171204,1413647919215906817),
    (24,1800452337108289740,1488963321701171204,1439747634277740546);



create table chat_rooms (rowid integer primary key,
    id text unique not null,
    type text not null,
    last_messaged_at integer not null,
    is_nsfw boolean not null,

    -- Group DM info
    created_at integer not null default 0,
    created_by_user_id integer not null default 0,
    name text not null default '',
    avatar_image_remote_url text not null default '',
    avatar_image_local_path text not null default ''
);

INSERT INTO chat_rooms VALUES
    (1,'1458284524761075714-1488963321701171204','ONE_TO_ONE',1686025129132,0,0,0,'','',''),
    (2,'1488963321701171204-1178839081222115328','ONE_TO_ONE',1686025129144,0,0,0,'','','');

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
INSERT INTO chat_room_participants VALUES
    (1,'1458284524761075714-1488963321701171204',1458284524761075714,1766595519000760325,0,0,0,0,0,0,''),
    (2,'1458284524761075714-1488963321701171204',1488963321701171204,1766595519000760325,1,0,0,0,1,0,'AT_END'),
    (3,'1488963321701171204-1178839081222115328',1488963321701171204,1665936253834578774,1,0,0,0,1,0,'TODO (what is it if there''s unreads?'),
    (4,'1488963321701171204-1178839081222115328',1178839081222115328,1665937253483614217,0,0,0,0,0,0,'');

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
INSERT INTO chat_messages VALUES
    (1,1663623062195957773,'1458284524761075714-1488963321701171204',1488963321701171204,1685473621419,'',0,'Yes helo',0),
    (2,1663623203644751885,'1458284524761075714-1488963321701171204',1458284524761075714,1685473655064,'',0,'Yeah i know who you are lol',0),
    (3,1665922180176044037,'1458284524761075714-1488963321701171204',1458284524761075714,1686021773787,'',1663623062195957773,'Yes?',0),
    (4,1665936253483614212,'1458284524761075714-1488963321701171204',1458284524761075714,1686025129132,'',0,replace('Check this out','\n',char(10)),1665509126737129472),
    (5,1665936253483614213,'1488963321701171204-1178839081222115328',1488963321701171204,1686025129140,'',0,'bruh1',0),
    (6,1665936253483614214,'1488963321701171204-1178839081222115328',1178839081222115328,1686025129141,'',0,'bruh2',0),
    (7,1665936253483614215,'1488963321701171204-1178839081222115328',1178839081222115328,1686025129142,'',1665936253483614214,'replying to bruh2',0),
    (8,1665936253483614216,'1488963321701171204-1178839081222115328',1488963321701171204,1686025129143,'',0,'This conversation is totally fake lol',0),
    (9,1665937253483614217,'1488963321701171204-1178839081222115328',1178839081222115328,1686025129144,'',0,'exactly',0),
    (36,1766248283901776125,'1458284524761075714-1488963321701171204',1458284524761075714,1709941380913,'',0,'',0),
    (15,1766255994668191902,'1458284524761075714-1488963321701171204',1458284524761075714,1709943219300,'',0,'You wrote this?',0),
    (46,1766595519000760325,'1458284524761075714-1488963321701171204',1458284524761075714,1710024168245,'',0,'This looks pretty good huh',0);


create table chat_message_reactions (rowid integer primary key,
    id integer unique not null check(typeof(id) = 'integer'),
    message_id integer not null,
    sender_id integer not null,
    sent_at integer not null,
    emoji text not null,
    foreign key(message_id) references chat_messages(id)
    foreign key(sender_id) references users(id)
);
INSERT INTO chat_message_reactions VALUES
    (1,1665914315742781440,1663623062195957773,1458284524761075714,1686019898732,'😂'),
    (2,1665936253487546456,1665936253483614216,1488963321701171204,1686063453455,'🤔'),
    (3,1665936253834578774,1665936253483614216,1178839081222115328,1686075343331,'🤔');

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
INSERT INTO chat_message_images VALUES(1,1766595500407459840,1766595519000760325,680,597,'https://ton.twitter.com/1.1/ton/data/dm/1766595519000760325/1766595500407459840/ML6pC79A.png','ML/ML6pC79A.png',0);

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
INSERT INTO chat_message_videos VALUES
    (1,1766248268416385024,1766248283901776125,500,280,'https://video.twimg.com/dm_video/1766248268416385024/vid/avc1/500x280/edFuZXtEVvem158AjvmJ3SZ_1DdG9cbSoW4fm6cDF1k.mp4?tag=1','ed/edFuZXtEVvem158AjvmJ3SZ_1DdG9cbSoW4fm6cDF1k.mp4','https://pbs.twimg.com/dm_video_preview/1766248268416385024/img/Ph7CCqISQxFE40Yy-uJAis-WiYhBbexFe_czkN5ytzI.jpg','Ph/Ph7CCqISQxFE40Yy-uJAis-WiYhBbexFe_czkN5ytzI.jpg',1980,0,0,0,0);

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
INSERT INTO chat_message_urls VALUES
    (1,1766255994668191902,'offline-twitter.com','https://offline-twitter.com/introduction/data-ownership-and-composability/','https://t.co/V3iiSYyrQx','Data ownership and composability','Data and Composability # What does it mean to own data? It means: You have a full copy of it It lasts until you decide to delete it You can do whatever you want with it, including opening it with...',0,0,0,0,'','',1,0,0);

create table follows(rowid integer primary key,
    follower_id integer not null,
    followee_id integer not null,
    unique(follower_id, followee_id),
    foreign key(follower_id) references users(id)
    foreign key(followee_id) references users(id)
);
create index if not exists index_follows_followee_id on follows (followee_id);
create index if not exists index_follows_follower_id on follows (follower_id);
insert into follows values
    (1, 1178839081222115328, 1488963321701171204), -- Mystery Grover follows Offline Twatter
    (2, 1032468021485293568, 1488963321701171204), -- Nemets follows Offline Twatter
    (3, 1488963321701171204, 1240784920831762433), -- Offline Twatter follows lukas
    (4, 1458284524761075714, 1240784920831762433), -- wispem-wantex follows lukas
    (5, 1458284524761075714, 358545917),           -- wispem-wantex follows cernovich
    (6, 1240784920831762433, 1458284524761075714); -- lukas follows wispem-wantex


create table notification_types (rowid integer primary key,
    name text not null unique
);
insert into notification_types(rowid, name) values
    (1, 'like'),
    (2, 'retweet'),
    (3, 'quote-tweet'),
    (4, 'reply'),
    (5, 'follow'),
    (6, 'mention'),
    (7, 'user is LIVE'),
    (8, 'poll ended'),
    (9, 'login'),
    (10, 'community pinned post'),
    (11, 'new recommended post');
create table notifications (rowid integer primary key,
    id text unique,
    type integer not null,
    sent_at integer not null,
    sort_index integer not null,
    user_id integer not null, -- user who received the notification

    action_user_id integer references users(id),  -- user who triggered the notification
    action_tweet_id integer references tweets(id), -- tweet associated with the notification
    action_retweet_id integer references retweets(retweet_id),

    has_detail boolean not null default 0,
    last_scraped_at not null default 0,

    foreign key(type) references notification_types(rowid)
    foreign key(user_id) references users(id)
);
INSERT INTO notifications VALUES
    (1,'FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4',1,1723494244885,1723494244885,1488963321701171204,249206992,1826778617705115868,NULL,1,1725067356270),
    (2,'FKncQJGVgAQAAAABSQ3bEaTgXL-G8wObqVY',9,1724112169072,1724112169072,1488963321701171204,NULL,NULL,NULL,0,-62135596800000),
    (3,'FKncQJGVgAQAAAABSQ3bEaTgXL_S11Ev36g',5,1722251072880,1724251072880,1488963321701171204,1032468021485293568,NULL,NULL,0,-62135596800000),
    (4,'FKncQJGVgAQAAAABSQ3bEaTgXL8VBxefepo',2,1724372973735,1724372973735,1488963321701171204,1178839081222115328,1826778617705115868,NULL,0,-62135596800000),
    (5,'FDzeDIfVUAIAAvsBiJONcqYgiLgXOolO9t0',6,-62135596800000,1725055975543,1488963321701171204,1178839081222115328,1826778617705115869,NULL,0,-62135596800000),
    (6,'FDzeDIfVUAIAAAABiJONcqaBFAzeN-n-Luw',1,1724604756351,1726604756351,1488963321701171204,1178839081222115328,NULL,1490135787124232223,0,-62135596800000);

create table notification_tweets (rowid integer primary key,
    notification_id not null references notifications(id),
    tweet_id not null references tweets(id),
    unique(notification_id, tweet_id)
);
insert into notification_tweets values
    (1, 'FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4', 1826778617705115868),
    (2, 'FKncQJGVgAQAAAABSQ3bEaTgXL8VBxefepo', 1826778617705115868),
    (3, 'FDzeDIfVUAIAAvsBiJONcqYgiLgXOolO9t0', 1826778617705115869),
    (4, 'FDzeDIfVUAIAAAABiJONcqaBFAzeN-n-Luw', 1507883724615999488);

create table notification_retweets (rowid integer primary key,
    notification_id not null references notifications(id),
    retweet_id not null references retweets(retweet_id),
    unique(notification_id, retweet_id)
);
insert into notification_retweets values
    (1, 'FDzeDIfVUAIAAAABiJONcqaBFAzeN-n-Luw', 1490135787124232223);

create table notification_users (rowid integer primary key,
    notification_id not null references notifications(id),
    user_id not null references users(id),
    unique(notification_id, user_id)
);
INSERT INTO notification_users VALUES
    (1,'FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4',249206992),
    (2,'FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4',1304281147074064385),
    (3,'FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4',1178839081222115328);


create table fake_user_sequence(latest_fake_id integer not null);
insert into fake_user_sequence values(0x4000000000000000);

create table database_version(rowid integer primary key,
    version_number integer not null unique
);
insert into database_version(version_number) values (31);

COMMIT;
