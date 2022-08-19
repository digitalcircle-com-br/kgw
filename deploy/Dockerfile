FROM alpine:latest

 RUN apk --no-cache add ca-certificates \
   && update-ca-certificates

COPY /deploy/kgw/main /app/main
WORKDIR /app
ENTRYPOINT ["./main"]