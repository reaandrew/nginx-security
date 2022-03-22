# nginx-security

This repo is exploring the different ways in which NGINX can be secured in Docker.  The majority of the methods are not specific to NGINX but NGINX is used as the subject.

1. Distroless image

This uses `gcr.io/distroless/base-debian10:nonroot` as the distroless base and `nginxinc/nginx-unprivileged` to copy the necessary files in the multi-stage build to run NGINX. 

To help find the minimal set of dependencies which need to be copy over, `ldd` was used which outputs something similar to:

```shell
root@c5eec8fbc239:/# ldd $(which nginx) 
        linux-vdso.so.1 (0x00007fffb4dd2000)
        libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007fed8f35d000)
        libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007fed8f33b000)
        libcrypt.so.1 => /lib/x86_64-linux-gnu/libcrypt.so.1 (0x00007fed8f300000)
        libpcre.so.3 => /lib/x86_64-linux-gnu/libpcre.so.3 (0x00007fed8f28d000)
        libssl.so.1.1 => /usr/lib/x86_64-linux-gnu/libssl.so.1.1 (0x00007fed8f1fa000)
        libcrypto.so.1.1 => /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 (0x00007fed8ef06000)
        libz.so.1 => /lib/x86_64-linux-gnu/libz.so.1 (0x00007fed8eee7000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007fed8ed22000)
        /lib64/ld-linux-x86-64.so.2 (0x00007fed8f4ab000)
```

This does not use the systemd daemon either, it runs NGINX in the foreground and relies on the daemon capabilities of Docker.  Not sure whether this is good, bad or of no concern yet.

Looking at the number of running processes both this distroless nginx version and the `nginxinc/nginx-unprivileged` have the same amount of processes.  The following is the output from `docker top <container-name>`

```shell
~/Development/nginx-security main !1 ❯ docker top nginx-secure
UID                 PID                 PPID                C                   STIME               TTY                 TIME                CMD
systemd+            109226              109205              0                   13:06               ?                   00:00:00            nginx: master process /usr/sbin/nginx -g daemon off;
systemd+            109259              109226              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109260              109226              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109261              109226              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109262              109226              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109263              109226              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109264              109226              0                   13:06               ?                   00:00:00            nginx: worker process

```

The `nginxinc/nginx-unprivileged` version

```shell
~/Development/nginx-security main !1 ❯ docker top happy_austin
UID                 PID                 PPID                C                   STIME               TTY                 TIME                CMD
systemd+            109461              109441              0                   13:06               ?                   00:00:00            nginx: master process nginx -g daemon off;
systemd+            109523              109461              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109524              109461              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109525              109461              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109526              109461              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109527              109461              0                   13:06               ?                   00:00:00            nginx: worker process
systemd+            109528              109461              0                   13:06               ?                   00:00:00            nginx: worker process
```

Looking at the size of the containers is a different story.  This distroless version is only ~28MB, which is over 100MB smaller than the `nginxinc/nginx-unprivileged`.

```shell
~/Development/nginx-security main ❯ docker images      
REPOSITORY                                 TAG       IMAGE ID       CREATED             SIZE
reaandrew/nginx-secure                     latest    c3230b0acdf8   About an hour ago   27.4MB
nginxinc/nginx-unprivileged                latest    b85bccd0d388   3 days ago          142MB
```

2. Readonly Filesystem


Really cool how they map to stdout and stderr the access and error log respecitively.

```shell
# forward request and error logs to docker log collector
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
```

[https://github.com/nginxinc/docker-nginx-unprivileged/blob/main/stable/debian/Dockerfile](https://github.com/nginxinc/docker-nginx-unprivileged/blob/main/stable/debian/Dockerfile)

Created a Makefile for convenience when testing so I can build, run, kill and show the logs really easily

```Makefile
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
```

The only other thing I needed to change was where the PID file got created.  Currently it is created directly in the `/tmp` directory and this would mean I would have to map the entire `/tmp` directory to make it work which is too much in terms of scope; I want more control than that.  So I followed a similar approach to the Dockerfile from `docker-nginx-unprivileged` and simply put the path inside a child directory of nginx.  This means I now can create a tempfs just for that directory and know exactly which directories will be writeable.

```Dockerfile
FROM nginxinc/nginx-unprivileged as build

RUN sed -i 's,/tmp/nginx.pid,/tmp/nginx/nginx.pid,' /etc/nginx/nginx.conf

...
```
