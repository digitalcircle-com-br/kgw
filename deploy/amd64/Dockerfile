FROM alpine:latest

 RUN apk --no-cache add ca-certificates \
   && update-ca-certificates

COPY /deploy/amd64/kgw /kgw/kgw
WORKDIR /
ENTRYPOINT ["/kgw/kgw"]