.PHONY: build
build:
	docker build -t reaandrew/nginx-secure .

.PHONY: run
run: build
	docker run --name nginx-secure --read-only \
		--mount type=tmpfs,destination=/tmp/proxy_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/client_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/fastcgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/uwsgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/scgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/nginx,tmpfs-size=1m \
		-d -p 8080:8080 -t reaandrew/nginx-secure

.PHONY: kill
kill:
	docker kill nginx-secure 2> /dev/null || :
	docker rm nginx-secure 2> /dev/null || :

.PHONY: logs
logs:
	docker logs nginx-secure
