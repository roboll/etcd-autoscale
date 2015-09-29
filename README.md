# etcd-autoscale-members

Get a list of autoscaling group members for etcd.

# About

`etcd` wants to know about the full cluster membership when it joins. As a result, running etcd in an autoscaling group can be kind of tricky. `etcd-autoscale-members` can help find a list of members you expect to be in the cluster, so that new members can join without needing the discovery service.

# Usage

Grab a github download, or build it yourself. There are some options.

    -group         - the autoscaling group name (required)
    -output-file   - the output file (defaults to /etc/sysconfig/etcd-members)
    -use-public-ip - use instance public ip (defaults to private ip)

The output file should contain a `source`-able or `EnvironmentFile`-able list of etcd members.

    ETCD_INITIAL_CLUSTER=i-xxxxxxxx=10.0.0.1,i-yyyyyyyy=10.0.0.2,i-zzzzzzzz=10.0.0.3

## Docker

`roboll/etcd-autoscale`
