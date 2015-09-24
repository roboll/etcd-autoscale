package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var groupName string
var outputFile string
var protocol string
var port string
var usePublicIP bool

func init() {
	flag.StringVar(&groupName, "group", "", "autoscaling group name")
	flag.StringVar(&outputFile, "output-file", "/etc/sysconfig/etcd-members", "output file: default /etc/sysconfig/etcd-members")
	flag.StringVar(&protocol, "protocol", "https://", "protocol to prefix, without :// ip: default https")
	flag.StringVar(&port, "port", "2379", "etcd port")
	flag.BoolVar(&usePublicIP, "use-public-ip", false, "use public ip: default false")
}

func main() {
	flag.Parse()
	if groupName == "" {
		println("Group name is required.")
		os.Exit(1)
	}
	if outputFile == "" {
		println("Output file is required.")
		os.Exit(1)
	}

	metadata := ec2metadata.New(&ec2metadata.Config{})
	region, err := metadata.Region()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	config := &aws.Config{
		Region: &region,
	}

	autoscale := autoscaling.New(config)
	groups, err := autoscale.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&groupName},
	})
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if len(groups.AutoScalingGroups) != 1 {
		println("Expected 1 autoscaling group.")
		os.Exit(1)
	}
	for _, group := range groups.AutoScalingGroups {
		instanceIds := []*string{}
		for _, instance := range group.Instances {
			instanceIds = append(instanceIds, instance.InstanceId)
		}

		aws := ec2.New(config)
		instances, err := aws.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: instanceIds,
		})
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}

		output := bytes.Buffer{}
		output.WriteString("ETCD_INITIAL_CLUSTER=")
		for residx, res := range instances.Reservations {
			for idx, instance := range res.Instances {
				if !(idx == 0 && residx == 0) {
					output.WriteString(",")
				}
				var ip *string
				if usePublicIP {
					ip = instance.PublicIpAddress
				} else {
					ip = instance.PrivateIpAddress
				}
				output.WriteString(fmt.Sprintf("%s://%s=%s:%s", protocol, *instance.InstanceId, *ip, port))
			}
		}

		file, err := os.Create(outputFile)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		output.WriteTo(file)
		err = file.Close()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
}
