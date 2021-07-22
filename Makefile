
build_client:
	cd ./client;npm run build;mv dist html;tar -czf html.tar.gz html;rm -rf html;scp ./html.tar.gz super_cnb:/home/one/super/html.tar.gz;rm -rf html.tar.gz;ssh super_cnb "cd super;tar -xzf html.tar.gz;rm html.tar.gz;exit"


build_client_en:
	cd ./client;npm run build_en;mv dist html;tar -czf html.tar.gz html;rm -rf html;scp ./html.tar.gz group:/home/one/super/html.tar.gz;rm -rf html.tar.gz;ssh group "cd super;tar -xzf html.tar.gz;rm html.tar.gz;exit"


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

monitor: build_server upload_cnb delete
	ssh super_cnb "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-monitor;exit;"

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

http_en: build_server upload_en delete
	ssh snapshot "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-http;exit;"

build_en: build_server upload_en delete
	ssh snapshot "cd super;gzip -d supergroup.gz;"

upload_en: build_server
	scp ./supergroup.gz snapshot:/home/one/super/supergroup.gz;
