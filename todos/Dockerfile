FROM alpine as build

RUN adduser -S todos todos

FROM gcr.io/distroless/base-debian10:nonroot

COPY ./todos /usr/bin/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group

USER todos
ENTRYPOINT ["/usr/bin/todos"]
