FROM alpine:edge
RUN apk --no-cache add alpine-sdk go
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT /entrypoint.sh