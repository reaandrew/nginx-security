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
