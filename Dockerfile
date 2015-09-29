###############################################################################
# https://github.com/roboll/etcd-autoscale
###############################################################################
FROM alpine:3.2

ADD target/etcd-autoscale-linux-amd64 /etcd-autoscale
ENTRYPOINT ["/etcd-autoscale"]
