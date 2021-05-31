
reload_mob:upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-http;sudo systemctl restart supergroup-blaze;sudo systemctl restart supergroup-distribute;sudo systemctl restart supergroup-assets-check"

assets_check:upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-assets-check"

add_client:
	ssh super_mob "cd super;./supergroup -service add_client"

http: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-http;exit;"

blaze: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;sudo systemctl restart supergroup-blaze;exit;"

build: build_server upload_mob delete
	ssh super_mob "cd super;rm supergroup;gzip -d supergroup.gz;"

upload_mob: build_server
	scp ./supergroup.gz super_mob:/home/one/super/supergroup.gz;

build_server:
	env GOOS=linux GOARCH=amd64 go build;gzip supergroup;

delete:
	rm -rf supergroup.gz

build_client:
	cd ./client;npm run build;mv dist html;tar -czf html.tar.gz html;rm -rf html;scp ./html.tar.gz super_mob:/home/one/super/html.tar.gz;rm -rf html.tar.gz;ssh super_mob "cd super;tar -xzf html.tar.gz;rm html.tar.gz;exit"
