package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func updateServiceImageCommand(awsSession *session.Session, cluster, service, container, image, tag string) {
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

	containerDefinition.Image = aws.String(image + ":" + tag)

	log.Printf("changed container image to %q", image+":"+tag)

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
