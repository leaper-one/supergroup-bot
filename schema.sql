create table if not exists client (
    client_id       varchar(36)                                            not null
    constraint client_pkey
    primary key,
    client_secret   varchar                                                not null,
    session_id      varchar(36)                                            not null,
    pin_token       varchar                                                not null,
    private_key     varchar                                                not null,
    pin             varchar(6),
    host            varchar                                                not null,
    asset_id        varchar(36)                                            not null,
    speak_status    smallint                 default 1                     not null,
    created_at      timestamp with time zone default now()                 not null,
    name            varchar                  default ''::character varying not null,
    description     varchar                  default ''::character varying not null,
    information_url varchar                  default ''::character varying
);

create table if not exists client_asset_level (
    client_id  varchar(36)                                             not null
    constraint client_asset_level_pkey
    primary key,
    fresh      varchar                  default '0'::character varying not null,
    senior     varchar                  default '0'::character varying not null,
    large      varchar                  default '0'::character varying not null,
    created_at timestamp with time zone default now()                  not null
);

create table if not exists client_replay (
    client_id       varchar(36)                            not null
    constraint client_replay_pkey
    primary key,
    join_msg        text                     default ''::text,
    join_url        varchar                  default ''::character varying,
    welcome         text                     default ''::text,
    limit_reject    text                     default ''::text,
    muted_reject    text                     default ''::text,
    category_reject text                     default ''::text,
    url_reject      text                     default ''::text,
    url_admin       text                     default ''::text,
    balance_reject  text                     default ''::text,
    updated_at      timestamp with time zone default now() not null
);

create table if not exists users (
    user_id         varchar(36)                            not null
    constraint users_pkey
    primary key,
    identity_number varchar                                not null
    constraint users_identity_number_key
    unique,
    full_name       varchar(512),
    avatar_url      varchar(1024),
    created_at      timestamp with time zone default now() not null
);

create table if not exists client_users (
    client_id      varchar(36)                            not null,
    user_id        varchar(36)                            not null,
    access_token   varchar(512),
    priority       smallint                 default 2     not null,
    is_async       boolean                  default true  not null,
    status         smallint                 default 0     not null,
    muted_time     varchar                  default ''::character varying,
    muted_at       timestamp with time zone default '1970-01-01 00:00:00+00'::timestamp with time zone,
    created_at     timestamp with time zone default now() not null,
    deliver_at     timestamp with time zone default now(),
    is_received    boolean                  default true,
    is_notice_join boolean                  default true,
    read_at        timestamp with time zone default now(),
    constraint client_users_pkey
    primary key (client_id, user_id)
);

create index if not exists client_user_idx
    on client_users (client_id);

create index if not exists client_user_priority_idx
    on client_users (client_id, priority);

create table if not exists client_asset_check (
    client_id  varchar(36)                            not null
    constraint client_asset_check_pkey
    primary key,
    asset_id   varchar(36),
    audience   varchar default '0'::character varying not null,
    fresh      varchar default '0'::character varying not null,
    senior     varchar default '0'::character varying not null,
    large      varchar default '0'::character varying not null,
    created_at timestamp with time zone               not null
);

create table if not exists assets (
    asset_id   varchar(36)   not null
    constraint assets_pkey
    primary key,
    chain_id   varchar(36)   not null,
    icon_url   varchar(1024) not null,
    symbol     varchar(128)  not null,
    name       varchar       not null,
    price_usd  varchar       not null,
    change_usd varchar       not null
);

create table if not exists messages (
    client_id        varchar(36)              not null,
    user_id          varchar(36)              not null,
    conversation_id  varchar(36)              not null,
    message_id       varchar(36)              not null,
    category         varchar,
    data             text,
    status           smallint                 not null,
    created_at       timestamp with time zone not null,
                                   quote_message_id varchar(36) default ''::character varying,
    constraint messages_pkey
    primary key (client_id, message_id)
);

create table if not exists distribute_messages (
    client_id         varchar(36)                               not null,
    user_id           varchar(36)                               not null,
    origin_message_id varchar(36)                               not null,
    message_id        varchar(36)                               not null,
    quote_message_id  varchar(36) default ''::character varying not null,
    level             smallint                                  not null,
    status            smallint    default 1                     not null,
    created_at        timestamp with time zone                  not null,
                                    data              text        default ''::text,
                                    category          varchar     default ''::character varying,
                                    representative_id varchar(36) default ''::character varying,
    conversation_id   varchar(36) default ''::character varying,
    shard_id          varchar(36) default ''::character varying,
    constraint distribute_messages_pkey
    primary key (client_id, user_id, origin_message_id)
);

create index if not exists distribute_messages_list_idx
    on distribute_messages (client_id, origin_message_id, level);

create index if not exists distribute_messages_all_list_idx
    on distribute_messages (client_id, shard_id, status, level, created_at);

create index if not exists distribute_messages_id_idx
    on distribute_messages (message_id);

create table if not exists swap (
    lp_asset      varchar(36)                                            not null
    constraint swap_pkey
    primary key,
    asset0        varchar(36)                                            not null,
    asset0_price  varchar                                                not null,
    asset0_amount varchar                  default ''::character varying not null,
    asset1        varchar(36)                                            not null,
    asset1_price  varchar                                                not null,
    asset1_amount varchar                  default ''::character varying not null,
    type          varchar(1)                                             not null,
    pool          varchar                                                not null,
    earn          varchar                                                not null,
    amount        varchar                                                not null,
    updated_at    timestamp with time zone default now()                 not null,
    created_at    timestamp with time zone default now()                 not null
);

create index if not exists swap_asset0_idx
    on swap (asset0);

create index if not exists swap_asset1_idx
    on swap (asset1);

create table if not exists exin_otc_asset (
    asset_id   varchar(36)                                                    not null
    constraint exin_otc_asset_pkey
    primary key,
    otc_id     varchar                                                        not null,
    price_usd  varchar                                                        not null,
    exchange   varchar                  default 'exchange'::character varying not null,
    buy_max    varchar                                                        not null,
    updated_at timestamp with time zone default now()                         not null
);

create table if not exists exin_local_asset (
    asset_id   varchar(36)                            not null,
    price      varchar                                not null,
    symbol     varchar                                not null,
    buy_max    varchar                                not null,
    updated_at timestamp with time zone default now() not null
);

create index if not exists exin_local_asset_id_idx
    on exin_local_asset (asset_id);

create table if not exists client_block_user (
    client_id  varchar(36)                            not null,
    user_id    varchar(36)                            not null,
    created_at timestamp with time zone default now() not null,
    constraint client_block_user_pkey
    primary key (client_id, user_id)
);

create table if not exists block_user (
    user_id    varchar(36)                            not null
    constraint block_user_pkey
    primary key,
    created_at timestamp with time zone default now() not null
);

create table if not exists client_asset_lp_check (
    client_id  varchar(36)                            not null,
    asset_id   varchar(36)                            not null,
    updated_at timestamp with time zone default now() not null,
    created_at timestamp with time zone default now() not null,
    constraint client_asset_lp_check_pkey
    primary key (client_id, asset_id)
);

create table if not exists broadcast (
    client_id  varchar(36)                                                                         not null,
    message_id varchar(36)                                                                         not null,
    status     smallint                 default 0                                                  not null,
    created_at timestamp with time zone default now()                                              not null,
    top_at     timestamp with time zone default '1970-01-01 00:00:00+00'::timestamp with time zone not null,
                             constraint broadcast_pkey
                             primary key (client_id, message_id)
);

create table if not exists activity (
    activity_index smallint                               not null
    constraint activity_pkey
    primary key,
    client_id      varchar(36)                            not null,
    status         smallint                 default 1,
    img_url        varchar(512)             default ''::character varying,
    expire_img_url varchar(512)             default ''::character varying,
    action         varchar(512)             default ''::character varying,
    start_at       timestamp with time zone               not null,
    expire_at      timestamp with time zone               not null,
    created_at     timestamp with time zone default now() not null
);

create table if not exists lives (
    live_id     varchar(36)                                                                         not null
    constraint lives_pkey
    primary key,
    client_id   varchar(36)                                                                         not null,
    img_url     varchar(512)             default ''::character varying,
    category    smallint                 default 1,
    title       varchar                                                                             not null,
    description varchar                                                                             not null,
    status      smallint                 default 1,
    created_at  timestamp with time zone default now()                                              not null,
    top_at      timestamp with time zone default '1970-01-01 00:00:00+00'::timestamp with time zone not null
);

create table if not exists live_replay (
    message_id varchar(36)                                            not null
    constraint live_replay_pkey
    primary key,
    client_id  varchar(36)                                            not null,
    live_id    varchar(36)              default ''::character varying not null,
    category   varchar                                                not null,
    data       varchar                                                not null,
    created_at timestamp with time zone default now()                 not null
);

create table if not exists live_data (
    live_id       varchar(36)                            not null
    constraint live_data_pkey
    primary key,
    read_count    integer                  default 0,
    deliver_count integer                  default 0,
    msg_count     integer                  default 0,
    user_count    integer                  default 0,
    start_at      timestamp with time zone default now() not null,
    end_at        timestamp with time zone default now() not null
);

