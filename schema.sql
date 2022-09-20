create table if not exists client (
	client_id varchar(36) not null primary key,
	client_secret varchar not null,
	session_id varchar(36) not null,
	pin_token varchar not null,
	private_key varchar not null,
	pin varchar(6),
	host varchar not null,
	asset_id varchar(36) not null,
	speak_status smallint default 1 not null,
	created_at timestamp with time zone default now() not null,
	name varchar default '' :: character varying not null,
	description varchar default '' :: character varying not null,
	icon_url varchar default '' :: character varying not null,
	owner_id varchar(36) default '' :: character varying not null,
	pay_amount varchar default '0' :: character varying,
	pay_status smallint default 0,
	identity_number varchar(11) default '' :: character varying,
	lang varchar default 'zh' :: character varying,
	admin_id varchar(36) default '' :: character varying not null
);

create table if not exists client_asset_level (
	client_id varchar(36) not null primary key,
	fresh varchar default '0' :: character varying not null,
	senior varchar default '0' :: character varying not null,
	large varchar default '0' :: character varying not null,
	created_at timestamp with time zone default now() not null,
	fresh_amount varchar default '0' :: character varying,
	large_amount varchar default '0' :: character varying
);

create table if not exists client_replay (
	client_id varchar(36) not null primary key,
	join_msg text default '' :: text,
	join_url varchar default '' :: character varying,
	welcome text default '' :: text,
	limit_reject text default '' :: text,
	muted_reject text default '' :: text,
	category_reject text default '' :: text,
	url_reject text default '' :: text,
	url_admin text default '' :: text,
	balance_reject text default '' :: text,
	updated_at timestamp with time zone default now() not null
);

create table if not exists users (
	user_id varchar(36) not null primary key,
	identity_number varchar not null unique,
	full_name varchar(512),
	avatar_url varchar(1024),
	created_at timestamp with time zone default now() not null,
	is_scam boolean default false,
	access_token varchar
);

create table if not exists client_users (
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	access_token varchar(512) default '' :: character varying not null,
	priority smallint default 2 not null,
	is_async boolean default true not null,
	status smallint default 0 not null,
	muted_time varchar default '' :: character varying,
	muted_at timestamp with time zone default '1970-01-01 00:00:00+00' :: timestamp with time zone,
	created_at timestamp with time zone default now() not null,
	deliver_at timestamp with time zone default now(),
	is_received boolean default true,
	is_notice_join boolean default true,
	read_at timestamp with time zone default now(),
	pay_status smallint default 1,
	pay_expired_at timestamp with time zone default '1970-01-01 00:00:00+00' :: timestamp with time zone,
	authorization_id varchar(36) default '' :: character varying,
	scope varchar(512) default '' :: character varying,
	private_key varchar(128) default '' :: character varying,
	ed25519 varchar(128) default '' :: character varying,
	primary key (client_id, user_id)
);

create index if not exists client_user_idx on client_users (client_id);

create index if not exists client_user_priority_idx on client_users (client_id, priority);

create table if not exists client_asset_check (
	client_id varchar(36) not null primary key,
	asset_id varchar(36),
	audience varchar default '0' :: character varying not null,
	fresh varchar default '0' :: character varying not null,
	senior varchar default '0' :: character varying not null,
	large varchar default '0' :: character varying not null,
	created_at timestamp with time zone not null
);

create table if not exists assets (
	asset_id varchar(36) not null primary key,
	chain_id varchar(36) not null,
	icon_url varchar(1024) not null,
	symbol varchar(128) not null,
	name varchar not null,
	price_usd varchar not null,
	change_usd varchar not null
);

create table if not exists messages (
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	conversation_id varchar(36) not null,
	message_id varchar(36) not null,
	category varchar,
	data text,
	status smallint not null,
	created_at timestamp with time zone not null,
	quote_message_id varchar(36) default '' :: character varying,
	primary key (client_id, message_id)
);

create table if not exists distribute_messages (
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	origin_message_id varchar(36) not null,
	message_id varchar(36) not null,
	quote_message_id varchar(36) default '' :: character varying not null,
	level smallint default 2 not null,
	status smallint default 1 not null,
	created_at timestamp with time zone default now() not null,
	data text default '' :: text,
	category varchar default '' :: character varying,
	representative_id varchar(36) default '' :: character varying,
	conversation_id varchar(36) default '' :: character varying,
	shard_id varchar(36) default '' :: character varying,
	primary key (client_id, user_id, origin_message_id)
);

create index if not exists distribute_messages_all_list_idx on distribute_messages (client_id, shard_id, status, level, created_at);

create index if not exists distribute_messages_id_idx on distribute_messages (message_id);

create index if not exists distribute_messages_list_idx on distribute_messages (client_id, origin_message_id, level);

create index if not exists remove_distribute_messages_id_idx on distribute_messages (status, created_at);

create table if not exists swap (
	lp_asset varchar(36) not null primary key,
	asset0 varchar(36) not null,
	asset0_price varchar not null,
	asset0_amount varchar default '' :: character varying not null,
	asset1 varchar(36) not null,
	asset1_price varchar not null,
	asset1_amount varchar default '' :: character varying not null,
	type varchar(1) not null,
	pool varchar not null,
	earn varchar not null,
	amount varchar not null,
	updated_at timestamp with time zone default now() not null,
	created_at timestamp with time zone default now() not null
);

create index if not exists swap_asset0_idx on swap (asset0);

create index if not exists swap_asset1_idx on swap (asset1);

create table if not exists exin_otc_asset (
	asset_id varchar(36) not null primary key,
	otc_id varchar not null,
	price_usd varchar not null,
	exchange varchar default 'exchange' :: character varying not null,
	buy_max varchar not null,
	updated_at timestamp with time zone default now() not null
);

create table if not exists exin_local_asset (
	asset_id varchar(36) not null,
	price varchar not null,
	symbol varchar not null,
	buy_max varchar not null,
	updated_at timestamp with time zone default now() not null
);

create index if not exists exin_local_asset_id_idx on exin_local_asset (asset_id);

create table if not exists client_block_user (
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	created_at timestamp with time zone default now() not null,
	primary key (client_id, user_id)
);

create index if not exists client_block_user_idx on client_block_user (client_id);

create table if not exists block_user (
	user_id varchar(36) not null primary key,
	created_at timestamp with time zone default now() not null,
	operator_id varchar(36) default '' :: character varying not null,
	memo varchar(255) default '' :: character varying not null
);

create table if not exists client_asset_lp_check (
	client_id varchar(36) not null,
	asset_id varchar(36) not null,
	updated_at timestamp with time zone default now() not null,
	created_at timestamp with time zone default now() not null,
	primary key (client_id, asset_id)
);

create table if not exists broadcast (
	client_id varchar(36) not null,
	message_id varchar(36) not null,
	status smallint default 0 not null,
	created_at timestamp with time zone default now() not null,
	top_at timestamp with time zone default '1970-01-01 00:00:00+00' :: timestamp with time zone not null,
	primary key (client_id, message_id)
);

create table if not exists activity (
	activity_index smallint not null primary key,
	client_id varchar(36) not null,
	status smallint default 1,
	img_url varchar(512) default '' :: character varying,
	expire_img_url varchar(512) default '' :: character varying,
	action varchar(512) default '' :: character varying,
	start_at timestamp with time zone not null,
	expire_at timestamp with time zone not null,
	created_at timestamp with time zone default now() not null
);

create table if not exists lives (
	live_id varchar(36) not null primary key,
	client_id varchar(36) not null,
	img_url varchar(512) default '' :: character varying,
	category smallint default 1,
	title varchar not null,
	description varchar not null,
	status smallint default 1,
	created_at timestamp with time zone default now() not null,
	top_at timestamp with time zone default '1970-01-01 00:00:00+00' :: timestamp with time zone not null
);

create table if not exists live_replay (
	message_id varchar(36) not null primary key,
	client_id varchar(36) not null,
	live_id varchar(36) default '' :: character varying not null,
	category varchar not null,
	data varchar not null,
	created_at timestamp with time zone default now() not null
);

create table if not exists live_data (
	live_id varchar(36) not null primary key,
	read_count integer default 0,
	deliver_count integer default 0,
	msg_count integer default 0,
	user_count integer default 0,
	start_at timestamp with time zone default now() not null,
	end_at timestamp with time zone default now() not null
);

create table if not exists live_play (
	live_id varchar(36) not null,
	user_id varchar not null,
	addr varchar default '' :: character varying not null,
	created_at timestamp with time zone default now() not null
);

create table if not exists daily_data (
	client_id varchar(36) not null,
	date date not null,
	users integer default 0 not null,
	active_users integer default 0 not null,
	messages integer default 0 not null,
	primary key (client_id, date)
);

create table if not exists snapshots (
	snapshot_id varchar(36) not null primary key,
	client_id varchar(36) not null,
	trace_id varchar(36) not null,
	user_id varchar(36) not null,
	asset_id varchar(36) not null,
	amount varchar not null,
	memo varchar default '' :: character varying,
	created_at timestamp with time zone not null
);

create table if not exists transfer_pendding (
	trace_id varchar(36) not null primary key,
	client_id varchar(36) not null,
	asset_id varchar(36) not null,
	opponent_id varchar(36) not null,
	amount varchar not null,
	memo varchar default '' :: character varying,
	status smallint default 1 not null,
	created_at timestamp with time zone not null
);

create table if not exists airdrop (
	airdrop_id varchar(36) not null,
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	asset_id varchar(36) not null,
	trace_id varchar(36) not null,
	amount varchar not null,
	status smallint default 1,
	created_at timestamp with time zone default now() not null,
	ask_amount varchar default '' :: character varying,
	primary key (airdrop_id, user_id)
);

create table if not exists client_white_url (
	client_id varchar(36) not null,
	white_url varchar default '' :: character varying not null,
	created_at timestamp with time zone default now(),
	primary key (client_id, white_url)
);

create table if not exists claim (
	user_id varchar(36) not null,
	date date default now() not null,
	ua varchar default '' :: character varying,
	addr varchar default '' :: character varying,
	client_id varchar default '' :: character varying,
	primary key (user_id, date)
);

create table if not exists power (
	user_id varchar(36) not null primary key,
	balance varchar default '0' :: character varying not null,
	lottery_times integer default 0 not null
);

create table if not exists power_record (
	power_type varchar(128) not null,
	user_id varchar(36) not null,
	amount varchar default '0' :: character varying not null,
	created_at timestamp default now() not null
);

create table if not exists lottery_record (
	lottery_id varchar(36) not null,
	user_id varchar(36) not null,
	asset_id varchar(36) not null,
	trace_id varchar(36) not null,
	snapshot_id varchar(36) default '' :: character varying not null,
	is_received boolean default false not null,
	amount varchar default '0' :: character varying not null,
	created_at timestamp default now() not null
);

create table if not exists guess (
	client_id varchar(36) not null,
	guess_id varchar(36) not null,
	asset_id varchar(36) not null,
	symbol varchar not null,
	price_usd varchar not null,
	rules varchar not null,
	explain varchar not null,
	start_time varchar not null,
	end_time varchar not null,
	start_at timestamp not null,
	end_at timestamp not null,
	created_at timestamp default now() not null
);

create table if not exists guess_record (
	guess_id varchar(36) not null,
	user_id varchar(36) not null,
	guess_type smallint not null,
	date date not null,
	result smallint default 0 not null,
	primary key (guess_id, user_id, date)
);

create table if not exists guess_result (
	asset_id varchar(36) not null,
	price varchar not null,
	date date default now() not null
);

create table if not exists login_log (
	user_id varchar(36) not null,
	client_id varchar(36) not null,
	addr varchar(255) not null,
	ua varchar(255) not null,
	updated_at timestamp default now() not null,
	ip_addr varchar default '' :: character varying not null,
	primary key (user_id, client_id)
);

create table if not exists client_member_auth (
	client_id varchar(36) not null,
	user_status smallint not null,
	plain_text boolean not null,
	lucky_coin boolean not null,
	plain_sticker boolean not null,
	plain_image boolean not null,
	plain_video boolean not null,
	plain_post boolean not null,
	plain_data boolean not null,
	plain_live boolean not null,
	plain_contact boolean not null,
	plain_transcript boolean not null,
	url boolean not null,
	updated_at timestamp default now() not null,
	app_card boolean default false,
	primary key (client_id, user_status)
);

create table if not exists properties (
	key varchar(512) not null primary key,
	value varchar(8192) not null,
	updated_at timestamp with time zone default now() not null
);

create table if not exists invitation (
	invitee_id varchar(36) not null primary key,
	inviter_id varchar(36) default '' :: character varying,
	client_id varchar(36) default '' :: character varying,
	invite_code varchar(6) not null unique,
	created_at timestamp with time zone default now()
);

create table if not exists invitation_power_record (
	invitee_id varchar(36) not null,
	inviter_id varchar(36) not null,
	amount varchar not null,
	created_at timestamp with time zone default now()
);

create table if not exists session (
	client_id varchar(36) not null,
	user_id varchar(36) not null,
	session_id varchar(36) not null,
	public_key varchar(128) not null,
	primary key (client_id, user_id, session_id)
);

create table if not exists liquidity_mining (
	mining_id varchar(36) not null primary key,
	title varchar not null,
	description varchar not null,
	faq varchar not null,
	join_tips varchar default '' :: character varying not null,
	join_url varchar default '' :: character varying not null,
	asset_id varchar(36) not null,
	client_id varchar(36) not null,
	first_time timestamp default now() not null,
	first_end timestamp default now() not null,
	daily_time timestamp default now() not null,
	daily_end timestamp default now() not null,
	reward_asset_id varchar(36) not null,
	first_amount varchar default '0' :: character varying not null,
	daily_amount varchar default '0' :: character varying not null,
	extra_asset_id varchar default '' :: character varying not null,
	extra_first_amount varchar default '0' :: character varying not null,
	extra_daily_amount varchar default '0' :: character varying not null,
	created_at timestamp default now() not null,
	first_desc varchar default '' :: character varying,
	daily_desc varchar default '' :: character varying,
	bg varchar default '' :: character varying
);

create table if not exists liquidity_mining_users (
	mining_id varchar(36) not null,
	user_id varchar(36) not null,
	created_at timestamp default now() not null,
	primary key (mining_id, user_id)
);

create table if not exists liquidity_mining_record (
	mining_id varchar(36) not null,
	record_id varchar(36) not null,
	user_id varchar(36) not null,
	asset_id varchar(36) not null,
	amount varchar default '0' :: character varying not null,
	profit varchar default '0' :: character varying not null,
	created_at timestamp default now() not null
);

create table if not exists liquidity_mining_tx (
	trace_id varchar(36) not null primary key,
	mining_id varchar(36) not null,
	record_id varchar(36) not null,
	user_id varchar(36) not null,
	asset_id varchar(36) not null,
	amount varchar default '0' :: character varying not null,
	status smallint default 0 not null,
	created_at timestamp default now() not null
);

create table if not exists lottery_supply (
	supply_id varchar(36) not null primary key,
	lottery_id varchar not null,
	asset_id varchar not null,
	amount varchar not null,
	client_id varchar not null,
	icon_url varchar not null,
	status smallint not null,
	created_at timestamp with time zone default now(),
	inventory integer default '-1' :: integer
);

create table if not exists lottery_supply_received (
	supply_id varchar(36) not null,
	user_id varchar(36) not null,
	trace_id varchar(36) not null,
	status smallint default 1 not null,
	created_at timestamp with time zone default now(),
	primary key (supply_id, user_id)
);

create table if not exists client_user_proxy (
	client_id varchar(36) not null,
	proxy_user_id varchar(36) not null,
	user_id varchar(36) not null,
	full_name varchar(255) not null,
	session_id varchar(36) not null,
	pin_token varchar not null,
	private_key varchar not null,
	status smallint default 1 not null,
	created_at timestamp with time zone default now(),
	primary key (client_id, proxy_user_id)
);

create index if not exists client_user_proxy_user_idx on client_user_proxy (client_id, user_id);

create table if not exists power_extra (
	client_id varchar(36) not null primary key,
	description varchar default '' :: character varying not null,
	multiplier varchar default '2' :: character varying not null,
	created_at timestamp with time zone default now(),
	start_at date default '1970-01-01' :: date,
	end_at date default '1970-01-01' :: date
);

create table if not exists trading_competition (
	competition_id varchar(36) not null primary key,
	client_id varchar(36) not null,
	asset_id varchar(36) not null,
	amount varchar not null,
	start_at date not null,
	end_at date not null,
	title varchar(255) not null,
	tips varchar(255) not null,
	rules varchar(255) not null,
	created_at timestamp default now() not null,
	reward varchar default '' :: character varying
);

create table if not exists user_snapshots (
	snapshot_id varchar(36) not null primary key,
	user_id varchar(36) not null,
	opponent_id varchar(36) not null,
	asset_id varchar(36) not null,
	amount varchar not null,
	opening_balance varchar not null,
	closing_balance varchar not null,
	source varchar not null,
	created_at timestamp default now() not null
);

create table if not exists trading_rank (
	competition_id varchar(36) not null,
	asset_id varchar(36) not null,
	user_id varchar(36) not null,
	amount varchar not null,
	updated_at timestamp default now() not null,
	primary key (competition_id, user_id)
);

create table if not exists voucher (
	code varchar(8) not null primary key,
	client_id varchar(36) default '' :: character varying,
	user_id varchar(36) default '' :: character varying,
	status smallint default 1 not null,
	updated_at timestamp with time zone default now() not null,
	created_at timestamp with time zone default now(),
	expired_at timestamp with time zone default (CURRENT_DATE + '7 days' :: interval)
);

create table if not exists client_menus (
	client_id varchar(36) not null,
	icon varchar(255) not null,
	name_zh varchar(255) not null,
	name_en varchar(255) not null,
	name_ja varchar(255) not null,
	url varchar(255) not null,
	idx integer default 0 not null,
	created_at timestamp default now() not null
);

create table if not exists channels (
	channel_id varchar(36) not null primary key,
	full_name varchar,
	avatar_url varchar,
	biography varchar,
	session_id varchar(36),
	pin_token varchar,
	private_key varchar,
	created_at timestamp with time zone default now()
);

create table if not exists articles (
	article_id varchar(36) not null primary key,
	channel_id varchar(36) constraint fk_articles_channel references channels,
	title varchar(255),
	cover varchar,
	abstract varchar,
	content text,
	status varchar,
	is_top boolean default false,
	updated_at timestamp with time zone default now(),
	created_at timestamp with time zone default now(),
	deleted_at timestamp with time zone,
	publisher_id varchar(36),
	published_at timestamp with time zone default now()
);

create index if not exists idx_articles_deleted_at on articles (deleted_at);

create table if not exists channel_managers (
	user_id varchar(36) not null,
	channel_id varchar(36) not null,
	created_at timestamp with time zone default now(),
	primary key (user_id, channel_id)
);

create table if not exists covers (
	channel_id varchar(36),
	cover_id varchar(36) not null primary key,
	url varchar(255),
	is_top boolean default false,
	top_at timestamp with time zone default now(),
	uploader_id varchar(36),
	created_at timestamp with time zone default now()
);

create index if not exists idx_cover_channel_id on covers (channel_id);

create table if not exists distribute_msgs (
	message_id varchar(36) not null primary key,
	recipient_id varchar(36),
	article_id varchar(36),
	status smallint
);

create index if not exists idx_distribute_msg_article_id on distribute_msgs (article_id);

create index if not exists idx_distribute_msg_status on distribute_msgs (status);

create table if not exists favorites (
	article_id varchar(36) not null,
	user_id varchar(36) not null,
	created_at timestamp with time zone default now(),
	primary key (article_id, user_id)
);

create index if not exists idx_favorite_user_id on favorites (user_id);

create index if not exists idx_favorite_article_id on favorites (article_id);

create table if not exists followings (
	channel_id varchar(36) not null,
	user_id varchar(36) not null,
	created_at timestamp with time zone default now(),
	primary key (channel_id, user_id)
);

create index if not exists idx_following_user_id on followings (user_id);

create index if not exists idx_following_channel_id on followings (channel_id);

create table if not exists likes (
	article_id varchar(36) not null constraint fk_likes_article references articles,
	user_id varchar(36) not null,
	created_at timestamp with time zone default now(),
	primary key (article_id, user_id)
);

create index if not exists idx_like_user_id on likes (user_id);

create index if not exists idx_like_article_id on likes (article_id);

create table if not exists views (
	article_id varchar(36) not null primary key,
	times bigint,
	updated_at timestamp with time zone default now()
);

create table if not exists view_records (
	article_id varchar(36),
	user_id varchar(36),
	ip varchar(255),
	ua varchar(255),
	created_at timestamp with time zone default now()
);

create table if not exists external_articles (
	hash text not null primary key,
	title text not null,
	"desc" text not null,
	author text not null,
	link text not null
);

create table if not exists report (
	reporter_id varchar(36) not null,
	reported_id varchar(36) not null,
	category varchar(4) not null,
	created_at timestamp default CURRENT_TIMESTAMP not null
);