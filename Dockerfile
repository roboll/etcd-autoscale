###############################################################################
# https://github.com/roboll/etcd-autoscale
###############################################################################
FROM alpine:3.2

RUN apk --update add ca-certificates && rm -rf /var/cache/apk/*

ADD target/etcd-autoscale-linux-amd64 /etcd-autoscale
ENTRYPOINT ["/etcd-autoscale"]
