FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /app

COPY server /app/server
RUN chmod +x /app/server

ENTRYPOINT ["./server", "--dir", "/data/sppdb", \
            "--contract", "0xcbd8084f8c759be749340bd20aaed48ec64860e6", \
            "--address", "0x5C9eDfaaC887552D6b521E38dAA3BFf1f645fD36"]
