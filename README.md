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
