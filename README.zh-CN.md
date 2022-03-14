# supergroup-bot
大群机器人的探索版，支持 http 服务和消息服务分离。

## 环境要求
- node
- go
- postgres
- redis > 6

## 安装
```shell
git clone https://github.com/MixinNetwork/supergroup-bot
```

## 后端配置

### 1. 配置信息(config.json)
字段介绍
| 字段             |                 介绍                  |
| :--------------- | :-----------------------------------: |
| port             |          运行的端口，如 7001          |
| lang             |        语言，目前只支持(zh/en)        |
| database         |          postgres数据库信息           |
| monitor          |      用于消息数据监控，可以为空       |
| qiniu            |  用于图文语音直播保存资源，可以为空   |
| redis_addr       |              redis的地址              |
| client_list      | 需要连接http和消息服务的client_id列表 |
| show_client_list |    在发现社群内显示的 client 列表     |
| luck_coin_app_id |             红包的app_id              |

> 目前红包的 app_id 只支持两个，中文版的`1ab1f241-b809-4790-bcfd-a1779bb1d313` 和英文版的`70b94e54-8f75-41f5-91e2-12522112ee71`


先配置好 `lang` | `database` | `redis_addr` 这三个字段。

### 2. 添加机器人大群(client.json)
字段介绍

| 字段                 |                        介绍                        |
| :------------------- | :------------------------------------------------: |
| client               |                大群的 keystore 信息                |
| client.client_secret |                     大群的密钥                     |
| client.host          |   web端打开大群的首页，如：http://localhost:8000   |
| client.name          |                      大群名字                      |
| client.description   |                 授权页面大群的简介                 |
| client.asset_id      |              大群的默认资产，可以为空              |
| level                |                      会员等级                      |
| replay.join_msg      |            未入群的成员会收到的默认回复            |
| replay.welcome       |            新入群的成员会收到的默认回复            |
| manager_list         | 管理员的 user_id 或 identity_number 列表，可以为空 |

先配置好 `client` 和 `replay`

### 3. 数据库配置
在第1步中配置的 `database` 的数据库里，先执行 `schema.sql`

### 4. 添加机器人大群到数据库中
准备好了 `client.json` 和 `config.json` 后
```shell
go run . -service add_client
```

### 5. 补充 config.json
再完善一下 `config.json` 中 `client_list` 和 `show_client_list`，这两个数组中，分别加入刚刚导入的 `client_id`

### 6. 启动服务
```shell
go run . # 启动 http 服务
go run . -service blaze & go run . -service create_message & go run . -service distribute_message # 启动消息服务
```

## 前端配置
以下文件介绍工作目录都是 /client
### 1. 编辑 `.umirc.dev.ts`
字段介绍

| 字段            |                              介绍                               |
| :-------------- | :-------------------------------------------------------------: |
| LANG            |                        目前仅支持 en/zh                         |
| MIXIN_BASE_URL  | 目前仅支持 https://mixin-api.zeromesh.net/https://api.mixin.one |
| RED_PACKET_ID   |                就是服务端配置的 luck_coin_app_id                |
| SERVER_URL      |       开发版请配置为服务端地址，如 http://localhost:7001        |
| LIVE_REPLAY_URL |                用于直播回放的视频前缀，可以为空                 |

> 若要部署上线，则请看后文部署篇

参考配置
```ts
import { defineConfig } from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.MIXIN_BASE_URL": "https://mixin-api.zeromesh.net",
    "process.env.RED_PACKET_ID": "1ab1f241-b809-4790-bcfd-a1779bb1d313",
    "process.env.SERVER_URL": "http://192.168.2.237:7001",
    "process.env.LIVE_REPLAY_URL": "",
  },
})
```

### 2. 安装依赖
```shell
npm install
```

### 3. 启动开发版
```shell
npm run start
```

> 注意：
> 1. 这里启动的浏览器地址，就是服务端需要配置的 `client.host` 的值，必须匹配才能正常启动
> 2. developers.mixin.one/dashboard 中，添加的机器人，也需要把 首页地址配置成 `client.host` 的值，然后验证地址是 `${client.host}/auth`

## 部署
### 1. 新建工作目录
```shell
mkdir supergroup-bot
```

### 2. 在服务端 `go build` 之后生成的静态文件放入瓷目录中

### 3. 在此目录中添加 `config.json`
并填充相应的信息

### 4. 域名配置
1. 需要开启两个域名，例如想开通 btc 的社群，那么可以创建两个二级域名 
> 1. btc.xxx.com
> 2. btc-api.xxx.com
2. 需要开启 `https` 
3. 在 `https://developers.mixin.one/dashboard` 中，将 首页地址改为 `https://btc.xxx.com` ，验证地址改为 `https://btc.xxx.com/auth`
4. 在 `nginx` 中配置 `btc-api.xxx.com` 到后端指定端口的服务
5. 在 `nginx` 中配置 `btc.xxx.com` 到前端指定工作目录下
6. 在 `/client/.umirc.prod.ts` 中修改 `SERVER_URL` 为 `xxx.com`
7. 此时准备工作完毕

### 5. 开始部署
1. 服务端启动服务
```shell
./supergroup # 启动 http 服务
./supergroup -service blaze & ./supergroup -service create_message & ./supergroup -service distribute_message # 启动消息服务
```
> 或者可以配置 `systemctl` 或者 `nohub` 来管理服务。 

2. 客户端启动服务
```shell
npm run build
```
然后将 dist 文件夹放到 `nginx` 指定 `btc.xxx.com` 的路径下。

3. 部署结束
