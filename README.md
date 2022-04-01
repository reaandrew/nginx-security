# nginx-security

This repo is exploring the different ways in which NGINX can be secured in Docker.  The majority of the methods are not specific to NGINX but NGINX is used as the subject.

## 1. **Distroless image**

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

## 2. **Readonly Filesystem**


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

## 3. **Disable inter-container communication**

**TL;DR;** 

Here I learn how the `--link` arg is now legacy and that `icc=false` is a recommended security practice as it disables communications on the default bridge network.  To allow containers to communicate with icc=false you need to use custom networks.

--

Disabling inter-container communication (icc) forces any containers to have explicit links with those it needs to communicate with.  This is a setting on the docker daemon itself and the setting can be applied in the `systemd` configuration.

To setup this example I have created a toy app called todos with very simple functionality.  I have split the project into two containers:

1. The Proxy - this is the existing work with nginx
2. The Todos App - this is a small go app to create, update and list todos

The idea is that there will be two container instances running - one for nginx and one for the todo app; a custom config is then supplied to nginx to proxy all calls to the todos app.  The nginx container instance will be the only one with any port mappings.

I wont detail all the code for the sample app or the nginx configuration, that can be found in the repository here [https://github.com/reaandrew/nginx-security](https://github.com/reaandrew/nginx-security)

### Setting up the two container instances

I have updated the Makefile to include the todos container. 

```Makefile
.PHONY: build
build:
	docker build -t reaandrew/nginx-secure ./proxy
	(cd todos && go build -o todos main.go)
	docker build -t reaandrew/todos-secure ./todos

.PHONY: run
run: build
	docker run --name todos-secure --read-only \
		-d -t reaandrew/todos-secure

	docker run --name nginx-secure --read-only \
		--mount type=tmpfs,destination=/tmp/proxy_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/client_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/fastcgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/uwsgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/scgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/nginx,tmpfs-size=1m \
		--link todos-secure:todos \
		-d -p 8080:8080 -t reaandrew/nginx-secure


.PHONY: kill
kill:
	docker kill nginx-secure 2> /dev/null || :
	docker rm nginx-secure 2> /dev/null || :
	docker kill todos-secure 2> /dev/null || :
	docker rm todos-secure 2> /dev/null || :


.PHONY: logs
logs:
	docker logs nginx-secure
	docker logs todos-secure
```

~~The important part is the addition of an explicit link `--link todos-secure:todos`.  This allows the nginx container to reference a host name of todos which you can see in the following nginx config snippet:~~

Later on I update this as this doesnt work when icc=false and is actually now a legacy approach.

```conf
...
    location / {
       proxy_pass   http://todos:8080;
    }
...
```

Also note that no ports have been exposed for the todos app, only the nginx container.

### Testing out the ICC switch

#### Before disabling ICC

To test this I am shelling into a busybox container and trying to consume the todos app directly, not via the nginx container.

Refresh the containers, kill and re-run.
```shell
make kill
make run
```

Review the containers which are running:

```shell
> docker ps
CONTAINER ID   IMAGE                    COMMAND                  CREATED          STATUS          PORTS                                       NAMES
da33f3a2bf52   reaandrew/nginx-secure   "/usr/sbin/nginx -g …"   43 seconds ago   Up 42 seconds   0.0.0.0:8080->8080/tcp, :::8080->8080/tcp   nginx-secure
2b7afded4cdf   reaandrew/todos-secure   "/usr/bin/todos"         43 seconds ago   Up 42 seconds                                               todos-secure
```

I want to start a new container which has a shell and some tools so I can test out inter container communication without using explcit links.  Before I do this I need to know what the IP is of the todo container.

```shell
> docker inspect todos-secure | jq -r '.[]|.NetworkSettings.Networks.bridge.IPAddress'
172.17.0.2
```

Now I have the IP I can start the other container, shell in and try to consume the todos list endpoint which should return an empty array.

```shell
> docker run -it --rm busybox
> wget -qO- 172.17.0.2:8080/todos
[]
```
Success, this shows we can communicate between containers, but I want to disable this.

#### After disabling ICC

To disable the inter-container communication you need to add an argument to the docker systemd file `/lib/systemd/system/docker.service` (the location may be different on your setup.)  The ExecStart line from that file should look like the following:

```unit file (systemd)
ExecStart=/usr/bin/dockerd --icc=false -H fd:// --containerd=/run/containerd/containerd.sock 
```

Now reload the systemd daemon since the content of the file has changed and restart the docker daemon.

```shell
> sudo systemctl daemon-reload
> sudo systemctl restart docker
```

Now when I try to consume the todos from the busybox I am expecting the process to hang since the inter-container communication has been disabled, but for the demo I will set a timeout to show it failing.

```shell
> make kill
> make run
> docker inspect todos-secure | jq -r '.[]|.NetworkSettings.Networks.bridge.IPAddress'
172.17.0.2
> docker run -it --rm busybox
> wget -T 2 -qO- 172.17.0.2:8080/todos
wget: download timed out
```

Now hitting the todos API from the host, the way it is intended to be used is also hanging!!

```shell
> curl -m 2 -v localhost:8080/todos
*   Trying 127.0.0.1:8080...
* Connected to localhost (127.0.0.1) port 8080 (#0)
> GET /todos HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.74.0
> Accept: */*
> 
* Operation timed out after 2003 milliseconds with 0 bytes received
* Closing connection 0
curl: (28) Operation timed out after 2003 milliseconds with 0 bytes received
```

**THIS DOESN'T WORK!!!**

### How to fix this and keep Inter-Container Communication disabled?

Turns out `--link` is now legacy along with the default bridge which I have used above.  The recomendation now is to use user-defined networks with docker including (for this example) user-defined bridges.  So the final steps are to create a custom bridge network, attach the two containers and see if the todos api can be consumed from the host.

Here is the update Makefile:

```Makefile
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
		-d -t reaandrew/todos-secure

	docker run --name nginx-secure --read-only \
		--mount type=tmpfs,destination=/tmp/proxy_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/client_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/fastcgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/uwsgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/scgi_temp,tmpfs-size=2m \
		--mount type=tmpfs,destination=/tmp/nginx,tmpfs-size=1m \
		--network appz \
		-d -p 8080:8080 -t reaandrew/nginx-secure


.PHONY: kill
kill:
	docker kill nginx-secure 2> /dev/null || :
	docker rm nginx-secure 2> /dev/null || :
	docker kill todos-secure 2> /dev/null || :
	docker rm todos-secure 2> /dev/null || :
	docker network rm appz


.PHONY: logs
logs:
	docker logs nginx-secure
	docker logs todos-secure

```

The changes to the Makefile include:

1. Create the network  Also update the kill target to remove the network
2. Remove the `--link` from the nginx container and instead attach both containers to the new `appz` network. 

The other change required is the hostname which nginx uses to reference the upstream service.  Before I defined this as `todos` but now this has changed to the name given to the container `todos-secure`.

The updated block in the nginx configuration file is now:

```conf
...
    location / {
       proxy_pass   http://todos-secure:8080;
    }
...
```

### Finally!

And now when the endpoint is hit from the host we get the correct response.

```shell
> curl -m 2 -v localhost:8080/todos
[]
```
