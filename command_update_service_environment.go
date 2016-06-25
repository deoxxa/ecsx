package main

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func updateServiceEnvironmentCommand(awsSession *session.Session, cluster, service, container string, variables []string) {
	svc := ecs.New(awsSession)

	s, err := getService(svc, cluster, service)
	if err != nil {
		panic(err)
	}

	log.Printf("found service %q (%q)", service, *(s.ServiceArn))

	describeTaskDefinitionOutput, err := svc.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: s.TaskDefinition,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("found task %q", *(s.TaskDefinition))

	containerDefinition := findContainerByName(describeTaskDefinitionOutput.TaskDefinition, container)
	if containerDefinition == nil {
		log.Fatalf("expected to find container named %q", container)
	}

	log.Printf("found container for %q", container)

	for _, v := range variables {
		a := strings.SplitN(v, "=", 2)
		if len(a) == 1 {
			a = append(a, "")
		}

		if p := findVariableByName(containerDefinition, a[0]); p != nil {
			log.Printf("changing variable %q from %q to %q", a[0], *p.Value, a[1])

			p.Value = aws.String(a[1])
		} else {
			log.Printf("adding variable %q as %q", a[0], a[1])

			containerDefinition.Environment = append(containerDefinition.Environment, &ecs.KeyValuePair{
				Name:  aws.String(a[0]),
				Value: aws.String(a[1]),
			})
		}
	}

	registerTaskDefinitionOutput, err := svc.RegisterTaskDefinition(&ecs.RegisterTaskDefinitionInput{
		Family:               describeTaskDefinitionOutput.TaskDefinition.Family,
		ContainerDefinitions: describeTaskDefinitionOutput.TaskDefinition.ContainerDefinitions,
		Volumes:              describeTaskDefinitionOutput.TaskDefinition.Volumes,
	})
	if err != nil {
		panic(err)
	}

	log.Printf("registered new task definition revision %d (%q)", *(registerTaskDefinitionOutput.TaskDefinition.Revision), *(registerTaskDefinitionOutput.TaskDefinition.TaskDefinitionArn))

	if _, err := svc.UpdateService(&ecs.UpdateServiceInput{
		Cluster:        aws.String(cluster),
		Service:        aws.String(service),
		TaskDefinition: registerTaskDefinitionOutput.TaskDefinition.TaskDefinitionArn,
	}); err != nil {
		panic(err)
	}

	log.Printf("applied service changes")

	if err := pollUntilTaskDeployed(svc, cluster, service, *(registerTaskDefinitionOutput.TaskDefinition.TaskDefinitionArn), printEvent); err != nil {
		panic(err)
	}

	log.Printf("deployed successfully")
}
