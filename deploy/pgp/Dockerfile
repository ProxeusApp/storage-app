FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app
COPY server /app/server

ENTRYPOINT ["./server", "-storageDir", "/data/pgpdb"]
