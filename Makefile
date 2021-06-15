
reload_cnb:upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-http;sudo systemctl restart supergroup-blaze;sudo systemctl restart supergroup-create-message;sudo systemctl restart supergroup-distribute;sudo systemctl restart supergroup-assets-check;sudo systemctl restart supergroup-swap"

reload_cnb_msg:upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-create-message;sudo systemctl restart supergroup-blaze;sudo systemctl restart supergroup-distribute;"

assets_check:upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-assets-check"

add_client:
	ssh super_cnb "cd super;./supergroup -service add_client"

http: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-http;exit;"

swap: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-swap;exit;"

blaze: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-blaze;exit;"

create: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-create-message;exit;"

distribute: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-distribute;exit;"

build: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;"

upload_cnb: build_server
	scp ./supergroup.gz super_cnb:/home/one/super/supergroup.gz;

build_server:
	env GOOS=linux GOARCH=amd64 go build;gzip supergroup;

delete:
	rm -rf supergroup.gz

build_client:
	cd ./client;npm run build;mv dist html;tar -czf html.tar.gz html;rm -rf html;scp ./html.tar.gz super_cnb:/home/one/super/html.tar.gz;rm -rf html.tar.gz;ssh super_cnb "cd super;tar -xzf html.tar.gz;rm html.tar.gz;exit"

reload_mob:upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-create-message;sudo systemctl restart supergroup-blaze;sudo systemctl restart supergroup-distribute"

reload_mob_blaze: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-blaze"

reload_mob_create: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-create-message"

reload_mob_distribute: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-distribute"

build_mob: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;"

upload_mob: build_server
	scp ./supergroup.gz super_mob:/home/one/super/supergroup.gz;

build_mob_ali: build_server upload_mob_ali delete
	ssh group_zh "cd super;gzip -d supergroup.gz;"

upload_mob_ali: build_server
	scp ./supergroup.gz group_zh:/home/one/super/supergroup.gz;

reload_service: build_server upload_group delete
	ssh group "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-create-message;sudo systemctl restart supergroup-blaze;sudo systemctl restart supergroup-distribute"

start_service: build_server upload_group delete
	ssh group "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl start supergroup-create-message;sudo systemctl start supergroup-blaze;sudo systemctl start supergroup-distribute"

upload_group: build_server
	scp ./supergroup.gz group:/home/one/super/supergroup.gz;

enable_service:
	scp ./supergroup*.service group:/home/one/super
	ssh group "cd super; sudo mv supergroup*.service /lib/systemd/system/;sudo systemctl enable supergroup-blaze.service;sudo systemctl enable supergroup-create-message.service;sudo systemctl enable supergroup-distribute.service;"
