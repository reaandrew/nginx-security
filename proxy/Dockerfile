FROM nginxinc/nginx-unprivileged as build

RUN sed -i 's,/tmp/nginx.pid,/tmp/nginx/nginx.pid,' /etc/nginx/nginx.conf

FROM gcr.io/distroless/base-debian10:nonroot

COPY --from=build /lib/x86_64-linux-gnu/libdl.so.2 /lib/x86_64-linux-gnu/libdl.so.2
COPY --from=build /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=build /lib/x86_64-linux-gnu/libz.so.1 /lib/x86_64-linux-gnu/libz.so.1
COPY --from=build /lib/x86_64-linux-gnu/libcrypt.so.1 /lib/x86_64-linux-gnu/libcrypt.so.1
COPY --from=build /lib/x86_64-linux-gnu/libpthread.so.0 /lib/x86_64-linux-gnu/libpthread.so.0

COPY --from=build /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

COPY --from=build /usr/lib/x86_64-linux-gnu/libssl.so.1.1 /usr/lib/x86_64-linux-gnu/libssl.so.1.1
COPY --from=build /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0 /usr/lib/x86_64-linux-gnu/libpcre2-8.so.0
COPY --from=build /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.1

COPY --from=build /usr/sbin/nginx /usr/sbin/nginx
COPY --from=build /var/log/nginx /var/log/nginx
COPY --from=build /etc/nginx /etc/nginx
COPY --from=build /usr/share/nginx/html /usr/share/nginx/html

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

COPY default.conf /etc/nginx/conf.d/


USER nginx
ENTRYPOINT ["/usr/sbin/nginx", "-g", "daemon off;"] 
