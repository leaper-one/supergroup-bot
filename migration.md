# 老群升级部署规则

### 第一阶段：社群所有者数据准备

1. web页面中入群的描述内容（文字）
2. 用户入群前的回复内容（文字）
3. 新用户首次入群的回复内容（文字）
4. 管理员列表
5. 提供 `client_id/session_id/private_key/client_secret/pin_token`
6. 导出 `users.sql` 和 `block.sql` 两个表
   > 描述： users.sql 需要包含字段及顺序如下：
   >
   >  user_id | access_token | subscribed_at | active_at
   >
   > block.sql 只需要包含字段 user_id 即可
7. 如果要开启社群持仓发言的话， 需要提供 对应的三个档的数量要求，以及 `asset_id`或不指定默认美元校验。

```sql
-- 导出 users.sql 表
create table tmp_view as select user_id,access_token,subscribed_at,active_at from users;
copy tmp_view to '/tmp/users.sql' with delimiter as '|';
drop table tmp_view;

-- 导出 block.sql 表
create table tmp_view as select user_id from blacklists;
copy tmp_view to '/tmp/block.sql' with delimiter as '|';
drop table tmp_view;
```

> 如果是远程的 psql 数据库
```shell
# 导出 users.sql 表
psql dbname -h hostname -U username -c "create table tmp_view as select user_id,access_token,subscribed_at,active_at from users;copy tmp_view to stdout with delimiter as '|';drop table tmp_view;" > users.sql

# 导出 block.sql 表
psql dbname -h hostname -U username -c "create table tmp_view as select user_id from blacklists;copy tmp_view to stdout with delimiter as '|';drop table tmp_view;" > block.sql
```


### 第二阶段：社群部署者准备

1. home_url 和 auth_url
2. 根据 [部署文档](./deploy.md)，将 home_url 和 api_url 添加入 nginx ，并开启 ssl。
3. 填充好 `client.json`
4. 将 `client.json` / `users.sql` / `block.sql` 放入 `supergroup` 同一目录下
5. 运行 `./supergroup -service migration`
   > 如果是开启了持仓检测的话，由于 exin 查询资产 api 的限制，每 10s 导入 100 个用户。过程可能相对比较慢。
6. 在要部署的服务器的 `config.json` 中添加 `client_id`
7. 在 `http` 的服务中 `config.json`，添加 `show_client_list`(如果有必要的话)

### 第三阶段：所有者和部署者配合操作

1. 所有者更换 `homeURL/authURL`
2. 所有者停止社群的消息服务
3. 所有者通知部署者已经完成上述操作
4. 部署者 运行 `restart http` `restart blaze/create_message/distribute_message`

### 总结： 社群拥有者升级社群所需要的操作：

1. 准备 `第一阶段` 的信息
2. 等待 `社群部署者` 的通知
3. 接收通知后，更换 `homeURL/authURL` 并停止社群的消息服务
4. 完成第三步后通知 `社群部署者`
5. 完成