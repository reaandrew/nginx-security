.PHONY: build
build:
	docker build -t reaandrew/nginx-secure ./proxy
	(cd todos && go build -o todos main.go)
	docker build -t reaandrew/todos-secure ./todos

.PHONY: run
run: build
	docker network create --driver bridge appz

	docker run --name todos-secure --read-only \
		--network appz \
		--restart=on-failure:3 \
		--ulimit nofile=4096 \
		--ulimit nproc=50 \
		--memory="1g" \
        --cpus="2" \
		-d -t reaandrew/todos-secure

	docker run --name nginx-secure --read-only \
		--mount type=tmpfs,destination=/tmp/proxy_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/client_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/fastcgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/uwsgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/scgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/nginx,tmpfs-size=1m \
		--network appz \
		--restart=on-failure:3 \
		--ulimit nofile=4096 \
		--ulimit nproc=50 \
		--memory="1g" \
		--cpus="2" \
		-d -p 8080:8080 -t reaandrew/nginx-secure


.PHONY: kill
kill:
	docker kill nginx-secure 2> /dev/null || :
	docker rm nginx-secure 2> /dev/null || :
	docker kill todos-secure 2> /dev/null || :
	docker rm todos-secure 2> /dev/null || :
	docker network rm appz 2> /dev/null || :


.PHONY: logs
logs:
	docker logs nginx-secure
	docker logs todos-secure

.PHONY: loadtest_1m
loadtest_1m:
	locust --headless \
		-u 60 \
		-r 2 \
		-t 1m \
		-f loadtest/locustfile.py \
		--host http://localhost:8080 \
		--html "loadtest_report_$(shell date +"%Y%m%d%H%M%S").html"