FROM alpine

WORKDIR /app
COPY server /app/server
COPY settings.json /app

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

#also redirect log to stdout
RUN ln -sf /dev/stdout /var/log/spp.log

#ENTRYPOINT ["./server", "--dir", "/data/sppdb", \
#            "--contract", "0x1cb227f8802f78b3810642f5f15f086f7c803604", \
#            "--address", "0x5C9eDfaaC887552D6b521E38dAA3BFf1f645fD36"]
ENTRYPOINT ["./server", "--dir", "/data/sppdb", \
            "--contract", "0xcbd8084f8c759be749340bd20aaed48ec64860e6", \
            "--address", "0x5C9eDfaaC887552D6b521E38dAA3BFf1f645fD36"]
