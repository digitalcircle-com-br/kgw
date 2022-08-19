FROM alpine:latest

 RUN apk --no-cache add ca-certificates \
   && update-ca-certificates

COPY /deploy/kgw /kgw/kgw
WORKDIR /kgw
ENTRYPOINT ["./kgw"]