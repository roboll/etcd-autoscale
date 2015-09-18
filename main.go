package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var groupName string
var outputFile string
var region string
var usePublicIP bool

func init() {
	flag.StringVar(&groupName, "group", "", "autoscaling group name")
	flag.StringVar(&outputFile, "output-file", "/etc/sysconfig/etcd-members", "output file: default /etc/sysconfig/etcd-members")
	flag.StringVar(&region, "aws-region", "", "aws region")
	flag.BoolVar(&usePublicIP, "use-public-ip", false, "use public ip: default false")
}

func main() {
	flag.Parse()
	if groupName == "" {
		log.Fatal("Group name is required.")
	}
	if outputFile == "" {
		log.Fatal("Output file is required.")
	}

	config := &aws.Config{}
	if region != "" {
		config.Region = &region
	}

	autoscale := autoscaling.New(config)
	groups, err := autoscale.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{&groupName},
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(groups.AutoScalingGroups) != 1 {
		fmt.Println(groups)
		log.Fatalf("Expected 1 autoscaling group, found %d.", len(groups.AutoScalingGroups))
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
			log.Fatal(err)
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
				output.WriteString(fmt.Sprintf("%s=%s", *instance.InstanceId, *ip))
			}
		}

		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatal(err)
		}
		output.WriteTo(file)
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}
}
