################
# Step 1: build executable
# suggestions from https://medium.com/@chemidy/create-the-smallest-and-secured-golang-docker-image-based-on-scratch-4752223b7324
FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git openssh-client ca-certificates tzdata && \
    update-ca-certificates && \
    mkdir -p $GOPATH/src/github.com/benkillin/NotificationsBot/

RUN adduser -D -g '' appuser

WORKDIR $GOPATH/src/github.com/benkillin/NotificationsBot/

COPY .git ./.git
COPY src ./src
COPY BotConfig.json .
COPY start.sh .
COPY bread.txt .

RUN go get -d -v ./src/cmd && \
    mkdir -p /opt/NotificationsBot/ && \
    mkdir -p /opt/NotificationsBot/bin/ && \
    mkdir -p /opt/NotificationsBot/logs/

RUN go build -o /opt/NotificationsBot/bin/NotificationsBot ./src/cmd && \
    cp BotConfig.json /opt/NotificationsBot/bin/ && \
    cp bread.txt /opt/NotificationsBot/bin/ && \
    cp start.sh /opt/NotificationsBot/bin/

##############################
# Step 2: build minimal image
FROM scratch

USER appuser

# without copying /bin/sh and the shared lib it uses, the cmd/entrypoint will not work.
COPY --from=builder /bin/sh /bin/
COPY --from=builder /lib/ld-musl-x86_64.so.1 /lib/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder --chown=appuser:appuser /opt /opt

WORKDIR /opt/NotificationsBot/bin/
ENTRYPOINT ["/opt/NotificationsBot/bin/NotificationsBot"]

#CMD /opt/NotificationsBot/bin/NotificationsBot
