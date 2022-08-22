FROM arm64v8/alpine:latest

 RUN apk --no-cache add ca-certificates \
   && update-ca-certificates

COPY /deploy/arm64/kgw /kgw/kgw
WORKDIR /
ENTRYPOINT ["/kgw/kgw"]