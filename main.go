package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var region string
var groupName string
var outputFile string
var protocol string
var port string
var usePublicIP bool
var hosts bool
var domain string

func init() {
	flag.StringVar(&region, "region", "", "aws region")
	flag.StringVar(&groupName, "group", "", "autoscaling group name")
	flag.StringVar(&outputFile, "output-file", "", "output file: default /etc/sysconfig/etcd-members")
	flag.StringVar(&protocol, "protocol", "http", "protocol to prefix, without :// ip: default http")
	flag.StringVar(&port, "port", "2379", "etcd port")
	flag.BoolVar(&usePublicIP, "use-public-ip", false, "use public ip: default false")
	flag.BoolVar(&hosts, "hosts", false, "write hosts config")
	flag.StringVar(&domain, "domain", "", "domain to append to ip")
}

func main() {
	flag.Parse()
	if groupName == "" {
		println("Group name is required.")
		os.Exit(1)
	}

	if region == "" {
		metadata := ec2metadata.New(&ec2metadata.Config{})
		r, err := metadata.Region()
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
		region = r
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

		hostOut := bytes.Buffer{}

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
				output.WriteString(fmt.Sprintf("%s=%s://%s.%s:%s", *instance.InstanceId, protocol, *ip, domain, port))
				hostname := "ip-" + strings.Replace(*ip, ".", "-", -1)
				hostOut.WriteString(fmt.Sprintf("%s %s.%s", *ip, hostname, domain))
				hostOut.WriteString("\n")
			}
		}

		if !hosts {
			if outputFile == "" {
				output.WriteTo(os.Stdout)
			} else {
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
			}
		} else {

			hostOut.WriteTo(os.Stdout)
		}
		/*
			hosts, err := os.Create("/etc/hosts")
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
		*/
		os.Exit(0)
	}
}
