FROM alpine

WORKDIR /app
COPY server /app/server

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#also redirect log to stdout
RUN ln -sf /dev/stdout /var/log/pgp.log

ENTRYPOINT ["./server", "-storageDir", "/data/pgpdb"]
