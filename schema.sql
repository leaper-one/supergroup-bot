CREATE TABLE activity (
	activity_index int2 NOT NULL,
	client_id varchar(36) NOT NULL,
	status int2 NULL DEFAULT 1,
	img_url varchar(512) NULL DEFAULT ''::character varying,
	expire_img_url varchar(512) NULL DEFAULT ''::character varying,
	"action" varchar(512) NULL DEFAULT ''::character varying,
	start_at timestamptz NOT NULL,
	expire_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT activity_pkey PRIMARY KEY (activity_index)
);


CREATE TABLE airdrop (
	airdrop_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	trace_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	status int2 NULL DEFAULT 1,
	created_at timestamptz NOT NULL DEFAULT now(),
	ask_amount varchar NULL DEFAULT ''::character varying,
	CONSTRAINT airdrop_pkey PRIMARY KEY (airdrop_id, user_id)
);


CREATE TABLE assets (
	asset_id varchar(36) NOT NULL,
	chain_id varchar(36) NOT NULL,
	icon_url varchar(1024) NOT NULL,
	symbol varchar(128) NOT NULL,
	"name" varchar NOT NULL,
	price_usd varchar NOT NULL,
	change_usd varchar NOT NULL,
	CONSTRAINT assets_pkey PRIMARY KEY (asset_id)
);


CREATE TABLE block_user (
	user_id varchar(36) NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT block_user_pkey PRIMARY KEY (user_id)
);


CREATE TABLE broadcast (
	client_id varchar(36) NOT NULL,
	message_id varchar(36) NOT NULL,
	status int2 NOT NULL DEFAULT 0,
	created_at timestamptz NOT NULL DEFAULT now(),
	top_at timestamptz NOT NULL DEFAULT '1970-01-01 08:00:00+08'::timestamp with time zone,
	CONSTRAINT broadcast_pkey PRIMARY KEY (client_id, message_id)
);


CREATE TABLE claim (
	user_id varchar(36) NOT NULL,
	"date" date NOT NULL DEFAULT now(),
	ua varchar NULL DEFAULT ''::character varying,
	addr varchar NULL DEFAULT ''::character varying,
	client_id varchar NULL DEFAULT ''::character varying,
	CONSTRAINT claim_pkey PRIMARY KEY (user_id, date)
);


CREATE TABLE client (
	client_id varchar(36) NOT NULL,
	client_secret varchar NOT NULL,
	session_id varchar(36) NOT NULL,
	pin_token varchar NOT NULL,
	private_key varchar NOT NULL,
	pin varchar(6) NULL,
	host varchar NOT NULL,
	asset_id varchar(36) NOT NULL,
	speak_status int2 NOT NULL DEFAULT 1,
	created_at timestamptz NOT NULL DEFAULT now(),
	"name" varchar NOT NULL DEFAULT ''::character varying,
	description varchar NOT NULL DEFAULT ''::character varying,
	icon_url varchar NOT NULL DEFAULT ''::character varying,
	owner_id varchar(36) NOT NULL DEFAULT ''::character varying,
	pay_amount varchar NULL DEFAULT '0'::character varying,
	pay_status int2 NULL DEFAULT 0,
	identity_number varchar(11) NULL DEFAULT ''::character varying,
	lang varchar NULL DEFAULT 'zh'::character varying,
	CONSTRAINT client_pkey PRIMARY KEY (client_id)
);


CREATE TABLE client_asset_check (
	client_id varchar(36) NOT NULL,
	asset_id varchar(36) NULL,
	audience varchar NOT NULL DEFAULT '0'::character varying,
	fresh varchar NOT NULL DEFAULT '0'::character varying,
	senior varchar NOT NULL DEFAULT '0'::character varying,
	"large" varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamptz NOT NULL,
	CONSTRAINT client_asset_check_pkey PRIMARY KEY (client_id)
);


CREATE TABLE client_asset_level (
	client_id varchar(36) NOT NULL,
	fresh varchar NOT NULL DEFAULT '0'::character varying,
	senior varchar NOT NULL DEFAULT '0'::character varying,
	"large" varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamptz NOT NULL DEFAULT now(),
	fresh_amount varchar NULL DEFAULT '0'::character varying,
	large_amount varchar NULL DEFAULT '0'::character varying,
	CONSTRAINT client_asset_level_pkey PRIMARY KEY (client_id)
);


CREATE TABLE client_asset_lp_check (
	client_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT client_asset_lp_check_pkey PRIMARY KEY (client_id, asset_id)
);


CREATE TABLE client_block_user (
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT client_block_user_pkey PRIMARY KEY (client_id, user_id)
);
CREATE INDEX client_block_user_idx ON client_block_user USING btree (client_id);


CREATE TABLE client_member_auth (
	client_id varchar(36) NOT NULL,
	user_status int2 NOT NULL,
	plain_text bool NOT NULL,
	lucky_coin bool NOT NULL,
	plain_sticker bool NOT NULL,
	plain_image bool NOT NULL,
	plain_video bool NOT NULL,
	plain_post bool NOT NULL,
	plain_data bool NOT NULL,
	plain_live bool NOT NULL,
	plain_contact bool NOT NULL,
	plain_transcript bool NOT NULL,
	url bool NOT NULL,
	updated_at timestamp NOT NULL DEFAULT now(),
	app_card bool NULL DEFAULT false,
	CONSTRAINT client_member_auth_pkey PRIMARY KEY (client_id, user_status)
);


CREATE TABLE client_replay (
	client_id varchar(36) NOT NULL,
	join_msg text NULL DEFAULT ''::text,
	join_url varchar NULL DEFAULT ''::character varying,
	welcome text NULL DEFAULT ''::text,
	limit_reject text NULL DEFAULT ''::text,
	muted_reject text NULL DEFAULT ''::text,
	category_reject text NULL DEFAULT ''::text,
	url_reject text NULL DEFAULT ''::text,
	url_admin text NULL DEFAULT ''::text,
	balance_reject text NULL DEFAULT ''::text,
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT client_replay_pkey PRIMARY KEY (client_id)
);


CREATE TABLE client_user_proxy (
	client_id varchar(36) NOT NULL,
	proxy_user_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	full_name varchar(255) NOT NULL,
	session_id varchar(36) NOT NULL,
	pin_token varchar NOT NULL,
	private_key varchar NOT NULL,
	status int2 NOT NULL DEFAULT 1,
	created_at timestamptz NULL DEFAULT now(),
	CONSTRAINT client_user_proxy_pkey PRIMARY KEY (client_id, proxy_user_id)
);
CREATE INDEX client_user_proxy_user_idx ON client_user_proxy USING btree (client_id, user_id);


CREATE TABLE client_users (
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	access_token varchar(512) NOT NULL DEFAULT ''::character varying,
	priority int2 NOT NULL DEFAULT 2,
	is_async bool NOT NULL DEFAULT true,
	status int2 NOT NULL DEFAULT 0,
	muted_time varchar NULL DEFAULT ''::character varying,
	muted_at timestamptz NULL DEFAULT '1970-01-01 08:00:00+08'::timestamp with time zone,
	created_at timestamptz NOT NULL DEFAULT now(),
	deliver_at timestamptz NULL DEFAULT now(),
	is_received bool NULL DEFAULT true,
	is_notice_join bool NULL DEFAULT true,
	read_at timestamptz NULL DEFAULT now(),
	pay_status int2 NULL DEFAULT 1,
	pay_expired_at timestamptz NULL DEFAULT '1970-01-01 08:00:00+08'::timestamp with time zone,
	CONSTRAINT client_users_pkey PRIMARY KEY (client_id, user_id)
);
CREATE INDEX client_user_idx ON client_users USING btree (client_id);
CREATE INDEX client_user_priority_idx ON client_users USING btree (client_id, priority);


CREATE TABLE client_white_url (
	client_id varchar(36) NOT NULL,
	white_url varchar NOT NULL DEFAULT ''::character varying,
	created_at timestamptz NULL DEFAULT now(),
	CONSTRAINT client_white_url_pkey PRIMARY KEY (client_id, white_url)
);


CREATE TABLE daily_data (
	client_id varchar(36) NOT NULL,
	"date" date NOT NULL,
	users int4 NOT NULL DEFAULT 0,
	active_users int4 NOT NULL DEFAULT 0,
	messages int4 NOT NULL DEFAULT 0,
	CONSTRAINT daily_data_pkey PRIMARY KEY (client_id, date)
);


CREATE TABLE distribute_messages (
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	origin_message_id varchar(36) NOT NULL,
	message_id varchar(36) NOT NULL,
	quote_message_id varchar(36) NOT NULL DEFAULT ''::character varying,
	"level" int2 NOT NULL DEFAULT 2,
	status int2 NOT NULL DEFAULT 1,
	created_at timestamptz NOT NULL DEFAULT now(),
	"data" text NULL DEFAULT ''::text,
	category varchar NULL DEFAULT ''::character varying,
	representative_id varchar(36) NULL DEFAULT ''::character varying,
	conversation_id varchar(36) NULL DEFAULT ''::character varying,
	shard_id varchar(36) NULL DEFAULT ''::character varying,
	CONSTRAINT distribute_messages_pkey PRIMARY KEY (client_id, user_id, origin_message_id)
);
CREATE INDEX distribute_messages_all_list_idx ON distribute_messages USING btree (client_id, shard_id, status, level, created_at);
CREATE INDEX distribute_messages_id_idx ON distribute_messages USING btree (message_id);
CREATE INDEX distribute_messages_list_idx ON distribute_messages USING btree (client_id, origin_message_id, level);
CREATE INDEX remove_distribute_messages_id_idx ON distribute_messages USING btree (status, created_at);


CREATE TABLE exin_local_asset (
	asset_id varchar(36) NOT NULL,
	price varchar NOT NULL,
	symbol varchar NOT NULL,
	buy_max varchar NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX exin_local_asset_id_idx ON exin_local_asset USING btree (asset_id);


CREATE TABLE exin_otc_asset (
	asset_id varchar(36) NOT NULL,
	otc_id varchar NOT NULL,
	price_usd varchar NOT NULL,
	exchange varchar NOT NULL DEFAULT 'exchange'::character varying,
	buy_max varchar NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT exin_otc_asset_pkey PRIMARY KEY (asset_id)
);


CREATE TABLE guess (
	client_id varchar(36) NOT NULL,
	guess_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	symbol varchar NOT NULL,
	price_usd varchar NOT NULL,
	rules varchar NOT NULL,
	"explain" varchar NOT NULL,
	start_time varchar NOT NULL,
	end_time varchar NOT NULL,
	start_at timestamp NOT NULL,
	end_at timestamp NOT NULL,
	created_at timestamp NOT NULL DEFAULT now()
);


CREATE TABLE guess_record (
	guess_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	guess_type int2 NOT NULL,
	"date" date NOT NULL,
	"result" int2 NOT NULL DEFAULT 0,
	CONSTRAINT guess_record_pkey PRIMARY KEY (guess_id, user_id, date)
);


CREATE TABLE guess_result (
	asset_id varchar(36) NOT NULL,
	price varchar NOT NULL,
	"date" date NOT NULL DEFAULT now()
);


CREATE TABLE invitation (
	invitee_id varchar(36) NOT NULL,
	inviter_id varchar(36) NULL DEFAULT ''::character varying,
	client_id varchar(36) NULL DEFAULT ''::character varying,
	invite_code varchar(6) NOT NULL,
	created_at timestamptz NULL DEFAULT now(),
	CONSTRAINT invitation_invite_code_key UNIQUE (invite_code),
	CONSTRAINT invitation_pkey PRIMARY KEY (invitee_id)
);


CREATE TABLE invitation_power_record (
	invitee_id varchar(36) NOT NULL,
	inviter_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	created_at timestamptz NULL DEFAULT now()
);


CREATE TABLE liquidity_mining (
	mining_id varchar(36) NOT NULL,
	title varchar NOT NULL,
	description varchar NOT NULL,
	faq varchar NOT NULL,
	join_tips varchar NOT NULL DEFAULT ''::character varying,
	join_url varchar NOT NULL DEFAULT ''::character varying,
	asset_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	first_time timestamp NOT NULL DEFAULT now(),
	first_end timestamp NOT NULL DEFAULT now(),
	daily_time timestamp NOT NULL DEFAULT now(),
	daily_end timestamp NOT NULL DEFAULT now(),
	reward_asset_id varchar(36) NOT NULL,
	first_amount varchar NOT NULL DEFAULT '0'::character varying,
	daily_amount varchar NOT NULL DEFAULT '0'::character varying,
	extra_asset_id varchar NOT NULL DEFAULT ''::character varying,
	extra_first_amount varchar NOT NULL DEFAULT '0'::character varying,
	extra_daily_amount varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamp NOT NULL DEFAULT now(),
	first_desc varchar NULL DEFAULT ''::character varying,
	daily_desc varchar NULL DEFAULT ''::character varying,
	bg varchar NULL DEFAULT ''::character varying,
	CONSTRAINT liquidity_mining_pkey PRIMARY KEY (mining_id)
);


CREATE TABLE liquidity_mining_record (
	mining_id varchar(36) NOT NULL,
	record_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0'::character varying,
	profit varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamp NOT NULL DEFAULT now()
);


CREATE TABLE liquidity_mining_tx (
	trace_id varchar(36) NOT NULL,
	mining_id varchar(36) NOT NULL,
	record_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0'::character varying,
	status int2 NOT NULL DEFAULT 0,
	created_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT liquidity_mining_tx_pkey PRIMARY KEY (trace_id)
);


CREATE TABLE liquidity_mining_users (
	mining_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT liquidity_mining_users_pkey PRIMARY KEY (mining_id, user_id)
);


CREATE TABLE live_data (
	live_id varchar(36) NOT NULL,
	read_count int4 NULL DEFAULT 0,
	deliver_count int4 NULL DEFAULT 0,
	msg_count int4 NULL DEFAULT 0,
	user_count int4 NULL DEFAULT 0,
	start_at timestamptz NOT NULL DEFAULT now(),
	end_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT live_data_pkey PRIMARY KEY (live_id)
);


CREATE TABLE live_play (
	live_id varchar(36) NOT NULL,
	user_id varchar NOT NULL,
	addr varchar NOT NULL DEFAULT ''::character varying,
	created_at timestamptz NOT NULL DEFAULT now()
);


CREATE TABLE live_replay (
	message_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	live_id varchar(36) NOT NULL DEFAULT ''::character varying,
	category varchar NOT NULL,
	"data" varchar NOT NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT live_replay_pkey PRIMARY KEY (message_id)
);


CREATE TABLE lives (
	live_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	img_url varchar(512) NULL DEFAULT ''::character varying,
	category int2 NULL DEFAULT 1,
	title varchar NOT NULL,
	description varchar NOT NULL,
	status int2 NULL DEFAULT 1,
	created_at timestamptz NOT NULL DEFAULT now(),
	top_at timestamptz NOT NULL DEFAULT '1970-01-01 08:00:00+08'::timestamp with time zone,
	CONSTRAINT lives_pkey PRIMARY KEY (live_id)
);


CREATE TABLE login_log (
	user_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	addr varchar(255) NOT NULL,
	ua varchar(255) NOT NULL,
	updated_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT login_log_pkey PRIMARY KEY (user_id, client_id)
);


CREATE TABLE lottery_record (
	lottery_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	trace_id varchar(36) NOT NULL,
	snapshot_id varchar(36) NOT NULL DEFAULT ''::character varying,
	is_received bool NOT NULL DEFAULT false,
	amount varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamp NOT NULL DEFAULT now()
);


CREATE TABLE lottery_supply (
	supply_id varchar(36) NOT NULL,
	lottery_id varchar NOT NULL,
	asset_id varchar NOT NULL,
	amount varchar NOT NULL,
	client_id varchar NOT NULL,
	icon_url varchar NOT NULL,
	status int2 NOT NULL,
	created_at timestamptz NULL DEFAULT now(),
	inventory int4 NULL DEFAULT '-1'::integer,
	CONSTRAINT lottery_supply_pkey PRIMARY KEY (supply_id)
);


CREATE TABLE lottery_supply_received (
	supply_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	trace_id varchar(36) NOT NULL,
	status int2 NOT NULL DEFAULT 1,
	created_at timestamptz NULL DEFAULT now(),
	CONSTRAINT lottery_supply_received_pkey PRIMARY KEY (supply_id, user_id)
);


CREATE TABLE messages (
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	conversation_id varchar(36) NOT NULL,
	message_id varchar(36) NOT NULL,
	category varchar NULL,
	"data" text NULL,
	status int2 NOT NULL,
	created_at timestamptz NOT NULL,
	quote_message_id varchar(36) NULL DEFAULT ''::character varying,
	CONSTRAINT messages_pkey PRIMARY KEY (client_id, message_id)
);


CREATE TABLE power (
	user_id varchar(36) NOT NULL,
	balance varchar NOT NULL DEFAULT '0'::character varying,
	lottery_times int4 NOT NULL DEFAULT 0,
	CONSTRAINT power_pkey PRIMARY KEY (user_id)
);


CREATE TABLE power_extra (
	client_id varchar(36) NOT NULL,
	description varchar NOT NULL DEFAULT ''::character varying,
	multiplier varchar NOT NULL DEFAULT '2'::character varying,
	created_at timestamptz NULL DEFAULT now(),
	start_at date NULL DEFAULT '1970-01-01'::date,
	end_at date NULL DEFAULT '1970-01-01'::date,
	CONSTRAINT power_extra_pkey PRIMARY KEY (client_id)
);


CREATE TABLE power_record (
	power_type varchar(128) NOT NULL,
	user_id varchar(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0'::character varying,
	created_at timestamp NOT NULL DEFAULT now()
);


CREATE TABLE properties (
	"key" varchar(512) NOT NULL,
	value varchar(8192) NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT properties_pkey PRIMARY KEY (key)
);


CREATE TABLE "session" (
	client_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	session_id varchar(36) NOT NULL,
	public_key varchar(128) NOT NULL,
	CONSTRAINT session_pkey PRIMARY KEY (client_id, user_id, session_id)
);


CREATE TABLE snapshots (
	snapshot_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	trace_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	memo varchar NULL DEFAULT ''::character varying,
	created_at timestamptz NOT NULL,
	CONSTRAINT snapshots_pkey PRIMARY KEY (snapshot_id)
);


CREATE TABLE swap (
	lp_asset varchar(36) NOT NULL,
	asset0 varchar(36) NOT NULL,
	asset0_price varchar NOT NULL,
	asset0_amount varchar NOT NULL DEFAULT ''::character varying,
	asset1 varchar(36) NOT NULL,
	asset1_price varchar NOT NULL,
	asset1_amount varchar NOT NULL DEFAULT ''::character varying,
	"type" varchar(1) NOT NULL,
	pool varchar NOT NULL,
	earn varchar NOT NULL,
	amount varchar NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now(),
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT swap_pkey PRIMARY KEY (lp_asset)
);
CREATE INDEX swap_asset0_idx ON swap USING btree (asset0);
CREATE INDEX swap_asset1_idx ON swap USING btree (asset1);


CREATE TABLE trading_competition (
	competition_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	start_at date NOT NULL,
	end_at date NOT NULL,
	title varchar(255) NOT NULL,
	tips varchar(255) NOT NULL,
	rules varchar(255) NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	reward varchar NULL DEFAULT ''::character varying,
	CONSTRAINT trading_competition_pkey PRIMARY KEY (competition_id)
);


CREATE TABLE trading_rank (
	competition_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	updated_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT trading_rank_pkey PRIMARY KEY (competition_id, user_id)
);


CREATE TABLE transfer_pendding (
	trace_id varchar(36) NOT NULL,
	client_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	opponent_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	memo varchar NULL DEFAULT ''::character varying,
	status int2 NOT NULL DEFAULT 1,
	created_at timestamptz NOT NULL,
	CONSTRAINT transfer_pendding_pkey PRIMARY KEY (trace_id)
);


CREATE TABLE user_snapshots (
	snapshot_id varchar(36) NOT NULL,
	user_id varchar(36) NOT NULL,
	opponent_id varchar(36) NOT NULL,
	asset_id varchar(36) NOT NULL,
	amount varchar NOT NULL,
	opening_balance varchar NOT NULL,
	closing_balance varchar NOT NULL,
	"source" varchar NOT NULL,
	created_at timestamp NOT NULL DEFAULT now(),
	CONSTRAINT user_snapshots_pkey PRIMARY KEY (snapshot_id)
);


CREATE TABLE users (
	user_id varchar(36) NOT NULL,
	identity_number varchar NOT NULL,
	full_name varchar(512) NULL,
	avatar_url varchar(1024) NULL,
	created_at timestamptz NOT NULL DEFAULT now(),
	is_scam bool NULL DEFAULT false,
	CONSTRAINT users_identity_number_key UNIQUE (identity_number),
	CONSTRAINT users_pkey PRIMARY KEY (user_id)
);