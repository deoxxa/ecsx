package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elb"
)

func surveyCommand(awsSession *session.Session) {
	svcECS := ecs.New(awsSession)
	svcELB := elb.New(awsSession)

	listClustersOutput, err := svcECS.ListClusters(&ecs.ListClustersInput{})
	if err != nil {
		panic(err)
	}

	describeClustersOutput, err := svcECS.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: listClustersOutput.ClusterArns,
	})
	if err != nil {
		panic(err)
	}

	for _, cluster := range describeClustersOutput.Clusters {
		fmt.Printf("Cluster: %s\n", *cluster.ClusterName)

		listServicesOutput, err := svcECS.ListServices(&ecs.ListServicesInput{
			Cluster: cluster.ClusterArn,
		})
		if err != nil {
			panic(err)
		}

		describeServicesOutput, err := svcECS.DescribeServices(&ecs.DescribeServicesInput{
			Cluster:  cluster.ClusterArn,
			Services: listServicesOutput.ServiceArns,
		})
		if err != nil {
			panic(err)
		}

		for _, service := range describeServicesOutput.Services {
			fmt.Printf("  Service: %s\n", *service.ServiceName)
			fmt.Printf("    Status: %s\n", *service.Status)
			fmt.Printf("    Desired: %d\n", *service.DesiredCount)
			fmt.Printf("    Running: %d\n", *service.RunningCount)
			fmt.Printf("    Pending: %d\n", *service.PendingCount)

			for _, lb := range service.LoadBalancers {
				describeLoadBalancersOutput, err := svcELB.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
					LoadBalancerNames: aws.StringSlice([]string{*lb.LoadBalancerName}),
				})
				if err != nil {
					panic(err)
				}

				d := describeLoadBalancersOutput.LoadBalancerDescriptions[0]

				fmt.Printf("    Load Balancer: %s\n", *lb.LoadBalancerName)
				fmt.Printf("      Source: %s (%d)\n", *lb.ContainerName, *lb.ContainerPort)
				fmt.Printf("      URL: %s\n", *d.DNSName)
			}

			describeTaskDefinitionOutput, err := svcECS.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
				TaskDefinition: service.TaskDefinition,
			})
			if err != nil {
				panic(err)
			}

			taskDefinition := describeTaskDefinitionOutput.TaskDefinition

			fmt.Printf("    Definition: %s (revision %d)\n", *taskDefinition.Family, *taskDefinition.Revision)

			for _, c := range taskDefinition.ContainerDefinitions {
				fmt.Printf("      Container: %s\n", *c.Image)
				fmt.Printf("        Read Only Root Filesystem: %v\n", c.ReadonlyRootFilesystem != nil && *c.ReadonlyRootFilesystem)
				fmt.Printf("        Ports:\n")
				for _, p := range c.PortMappings {
					fmt.Printf("          %s %d -> %d\n", *p.Protocol, *p.ContainerPort, *p.HostPort)
				}
				fmt.Printf("        Variables:\n")
				for _, p := range c.Environment {
					fmt.Printf("          %s = %q\n", *p.Name, *p.Value)
				}
			}
		}
	}
}
