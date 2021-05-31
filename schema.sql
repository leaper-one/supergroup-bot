
-- 机器人信息
CREATE TABLE IF NOT EXISTS client {
  client_id          VARCHAR(36) NOT NULL PRIMARY KEY,
  client_secret      VARCHAR NOT NULL,
  session_id         VARCHAR(36) NOT NULL,
  pin_token          VARCHAR NOT NULL,
  private_key        VARCHAR NOT NULL,
  pin                VARCHAR(6) DEFAULT '',
  name               VARCHAR NOT NULL,
  description        VARCHAR NOT NULL,
  host               VARCHAR NOT NULL, -- 前端部署的 host
  asset_id           VARCHAR(36) NOT NULL,
  information_url    VARCHAR DEFAULT '',
  speak_status       SMALLINT NOT NULL DEFAULT 1, -- 1 正常发言 2 持仓发言
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
}

-- 用户的持仓等级
CREATE TABLE IF NOT EXISTS client_asset_level (
  client_id          VARCHAR(36) PRIMARY KEY,
  fresh              VARCHAR NOT NULL DEFAULT '0',
  senior             VARCHAR NOT NULL DEFAULT '0',
  large              VARCHAR NOT NULL DEFAULT '0',
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS client_replay (
  client_id          VARCHAR(36) NOT NULL PRIMARY KEY,
  join_msg           TEXT DEFAULT '', -- 入群前的内容
  join_url           VARCHAR DEFAULT '', -- 入群前发送的url

  welcome            TEXT DEFAULT '', -- 入群时发送的内容

  limit_reject       TEXT DEFAULT '', -- 1分钟发言次数超过限额
  muted_reject       TEXT DEFAULT '', -- 被禁言

  category_reject    TEXT DEFAULT '', -- 类型 被拦截消息

  url_reject         TEXT DEFAULT '', -- 链接被拦截消息
  url_admin          TEXT DEFAULT '', -- 转发给管理员的url消息

  balance_reject     TEXT DEFAULT '', -- 不满足持仓，不能发言
  updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 用户信息
CREATE TABLE IF NOT EXISTS users (
  user_id           VARCHAR(36) PRIMARY KEY,
  identity_number   VARCHAR NOT NULL UNIQUE,
  full_name         VARCHAR(512),
  avatar_url        VARCHAR(1024),
  created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 机器人用户信息表
CREATE TABLE IF NOT EXISTS client_users (
  client_id          VARCHAR(36),
  user_id            VARCHAR(36),
  access_token       VARCHAR(512),
  priority           SMALLINT NOT NULL DEFAULT 2, -- 1 优先级高 2 优先级低 3 补发中
  is_async           BOOLEAN NOT NULL DEFAULT true,
  status             SMALLINT NOT NULL DEFAULT 0, -- 0 未入群 1 观众 2 入门 3 资深 5 大户 8 嘉宾 9 管理
  muted_time         VARCHAR DEFAULT '',
  muted_at           TIMESTAMP WITH TIME ZONE,
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY (client_id, user_id)
);
CREATE INDEX client_user_idx ON client_users(client_id);
CREATE INDEX client_user_priority_idx ON client_users(client_id, priority);

-- 机器人 持仓级别表
CREATE TABLE IF NOT EXISTS client_asset_check (
  client_id          VARCHAR(36) PRIMARY KEY,
  asset_id           VARCHAR(36),
  audience           VARCHAR NOT NULL DEFAULT '0',
  fresh              VARCHAR NOT NULL DEFAULT '0',
  senior             VARCHAR NOT NULL DEFAULT '0',
  large              VARCHAR NOT NULL DEFAULT '0',
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL
);

-- 机器人 lp token 换算表
CREATE TABLE IF NOT EXISTS client_asset_lp_check (
  client_id          VARCHAR(36),
  asset_id           VARCHAR(36),
  price_usd          VARCHAR,
  updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 资产信息
CREATE TABLE IF NOT EXISTS assets (
    asset_id            VARCHAR(36) NOT NULL PRIMARY KEY,
    chain_id			VARCHAR(36) NOT NULL,
    icon_url            VARCHAR(1024) NOT NULL,
    symbol              VARCHAR(128) NOT NULL,
    name				VARCHAR NOT NULL,
    price_usd           VARCHAR NOT NULL,
    change_usd          VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
  conversation_id     VARCHAR(36) NOT NULL,
  message_id          VARCHAR(36) NOT NULL,
  quote_message_id    VARCHAR(36) NOT NULL DEFAULT '',
  category            VARCHAR,
  data                TEXT,
  status              SMALLINT NOT NULL, -- 1 pending 2 privilege 3 normal 4 finished
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY(client_id, message_id)
);

-- 分发的消息
CREATE TABLE IF NOT EXISTS distribte_messages (
  client_id             VARCHAR(36) NOT NULL,
  user_id               VARCHAR(36) NOT NULL,
  message_id            VARCHAR(36) NOT NULL,
  origin_message_id     VARCHAR(36) NOT NULL,
  level                 SMALLINT NOT NULL, -- 1 高优先级 2 低优先级 3 单独队列
  status                SMALLINT NOT NULL, -- 1 待分发 2 已分发
  created_at            TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY(client_id, user_id, message_id)
);


CREATE TABLE IF NOT EXISTS swap (
  lp_asset            VARCHAR(36) NOT NULL PRIMARY KEY, -- lpToken asset_id
  asset0              VARCHAR(36) NOT NULL, -- asset0 asset_id
  asset0_price        VARCHAR NOT NULL, -- asset0 价格
  asset0_amount       VARCHAR NOT NULL DEFAULT '', -- asset0 数量
  asset1              VARCHAR(36) NOT NULL, -- asset1 asset_id
  asset1_price        VARCHAR NOT NULL, -- asset1 价格
  asset1_amount       VARCHAR NOT NULL DEFAULT '', -- asset1 数量
  type                VARCHAR(1) NOT NULL, -- 0 exinswap交易 1 4swap交易 2 ExinOne交易 3 ExinLocal交易
  pool                VARCHAR NOT NULL, -- 资金池总量
  earn                VARCHAR NOT NULL, -- 24小时年化收益率
  amount              VARCHAR NOT NULL, -- 24小时总交易量
  updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(), -- 更新时间
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- 创建时间
);
CREATE INDEX swap_asset0_idx ON swap(asset0);
CREATE INDEX swap_asset1_idx ON swap(asset1);

CREATE TABLE IF NOT EXISTS exin_otc_asset (
  asset_id            VARCHAR(36) NOT NULL PRIMARY KEY,
  otc_id              VARCHAR NOT NULL,
  price_usd           VARCHAR NOT NULL,
  exchange            VARCHAR NOT NULL DEFAULT 'exchange',
  buy_max             VARCHAR NOT NULL,
  updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- 更新时间
);

CREATE TABLE IF NOT EXISTS exin_local_asset (
  asset_id            VARCHAR(36) NOT NULL,
  price               VARCHAR NOT NULL,
  symbol              VARCHAR NOT NULL,
  buy_max             VARCHAR NOT NULL,
  updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- 更新时间
);
CREATE INDEX exin_local_asset_id_idx ON exin_local_asset(asset_id);

CREATE TABLE IF NOT EXISTS client_block_user (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY (client_id,user_id)
);

CREATE TABLE IF NOT EXISTS block_user (
  user_id             VARCHAR(36) NOT NULL PRIMARY KEY,
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

INSERT INTO block_user(user_id) VALUES ('3fa99d76-7606-4151-8cab-c771e2398693');
INSERT INTO block_user(user_id) VALUES ('9621b90b-4db2-4432-9c16-d74a66c7b027');
INSERT INTO block_user(user_id) VALUES ('b495cfce-2bb3-4cf3-9139-40f0da7378e2');
INSERT INTO block_user(user_id) VALUES ('99b7356b-c51f-4bb0-9197-6b20ef04198b');



