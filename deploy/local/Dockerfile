FROM alpine:latest

 RUN apk --no-cache add ca-certificates \
   && update-ca-certificates

COPY /deploy/local/kgw /kgw/kgw
WORKDIR /
ENTRYPOINT ["/kgw/kgw"]