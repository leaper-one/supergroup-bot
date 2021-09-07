[中文版](README.zh-CN.md)
# supergroup-bot
The exploration version of the large group of robots supports the separation of http service and message service.

## Environmental requirements
-node
-go
-postgres
-redis

## Install
```shell
git clone https://github.com/MixinNetwork/supergroup-bot
```

## Backend configuration

### 1. Configuration information (config.json)
Field introduction
| Field | Introduction |
| :--------------- | :------------------------------- ----: |
| port | The running port, such as 7001 |
| lang | Language, currently only supports (zh/en) |
| database | postgres database information |
| monitor | Used for message data monitoring, can be empty |
| qiniu | Used to save resources for live broadcast of graphics, text and voice, can be empty |
| redis_addr | redis address |
| client_list | A list of client_ids that need to connect to http and message services |
| show_client_list | Client list displayed in the discovery community |
| luck_coin_app_id | red envelope app_id |

> Currently the red envelope app_id only supports two, the Chinese version of `1ab1f241-b809-4790-bcfd-a1779bb1d313` and the English version of `70b94e54-8f75-41f5-91e2-12522112ee71`


First configure the three fields `lang` | `database` | `redis_addr`.

### 2. Add a large group of robots (client.json)
Field introduction

| Field | Introduction |
| :------------------- | :--------------------------- ---------------------: |
| client | Large group of keystore information |
| client.client_secret | Large group of secret keys |
| client.host | Open the homepage of the large group on the web side, such as: http://localhost:8000 |
| client.name | Group name |
| client.description | Introduction to the authorization page group |
| client.asset_id | The default asset of the large group, can be empty |
| level | Member level |
| replay.join_msg | Default replies received by members who have not joined the group |
| replay.welcome | The default reply that new members will receive |
| manager_list | The user_id or identity_number list of the administrator, can be empty |

Configure `client` and `replay` first

### 3. Database configuration
In the database of `database` configured in step 1, first execute `schema.sql`

### 4. Add a large group of robots to the database
After preparing `client.json` and `config.json`
```shell
go run. -service add_client
```

### 5. Supplement config.json
Improve the `client_list` and `show_client_list` in `config.json`. In these two arrays, add the newly imported `client_id`

### 6. Start the service
```shell
go run. #Start http service
go run. -service blaze & go run. -service create_message & go run. -service distribute_message # Start message service
```

## Frontend configuration
The following files introduce the working directory is /client
### 1. Edit `.umirc.dev.ts`
Field introduction

| Field | Introduction |
| :-------------- | :-------------------------------- -----------------------------: |
| LANG | Currently only supports en/zh |
| MIXIN_BASE_URL | Currently only supports https://mixin-api.zeromesh.net/https://api.mixin.one |
| RED_PACKET_ID | is the luck_coin_app_id configured on the server side |
| SERVER_URL | Please configure the development version as the server address, such as http://localhost:7001 |
| LIVE_REPLAY_URL | The video prefix used for live playback, can be empty |

> If you want to deploy online, please see the deployment article later

Reference configuration
```ts
import {defineConfig} from "umi"

export default defineConfig({
  define: {
    "process.env.LANG": "zh",
    "process.env.MIXIN_BASE_URL": "https://api.mixin.one",
    "process.env.RED_PACKET_ID": "70b94e54-8f75-41f5-91e2-12522112ee71",
    "process.env.SERVER_URL": "http://192.168.2.237:7001",
    "process.env.LIVE_REPLAY_URL": "",
  },
})
```

### 2. Installation dependencies
```shell
npm install
```

### 3. Start the development version
```shell
npm run start
```

> Note:
> 1. The browser address started here is the value of `client.host` that the server needs to configure, and it must match to start normally
> 2. In developers.mixin.one/dashboard, the added robot also needs to configure the home page address to the value of `client.host`, and then verify that the address is `${client.host}/auth`

## Deployment
### 1. Create a new working directory
```shell
mkdir supergroup-bot
```

### 2. The static files generated after `go build` on the server are placed in the porcelain directory

### 3. Add `config.json` to this directory
And fill in the corresponding information

### 4. Domain name configuration
1. You need to open two domain names. For example, if you want to open a btc community, you can create two second-level domain names
> 1. btc.xxx.com
> 2. btc-api.xxx.com
2. Need to open `https`
3. In `https://developers.mixin.one/dashboard`, change the homepage address to `https://btc.xxx.com` and the verification address to `https://btc.xxx.com/ auth`
4. Configure `btc-api.xxx.com` in `nginx` to the back-end specified port service
5. Configure `btc.xxx.com` in `nginx` to the specified working directory of the front end
6. Modify `SERVER_URL` to `xxx.com` in `/client/.umirc.prod.ts`
7. The preparations are now complete

### 5. Start deployment
1. Server start service
```shell
./supergroup # start http service
./supergroup -service blaze & ./supergroup -service create_message & ./supergroup -service distribute_message # start message service
```
> Or you can configure `systemctl` or `nohub` to manage services.

2. Client start service
```shell
npm run build
```
Then put the dist folder in the path specified by `btc.xxx.com` by `nginx`.

3. End of deployment