package main

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
)

func scaleServiceCommand(awsSession *session.Session, cluster, service string, count int64) {
	svc := ecs.New(awsSession)

	s, err := getService(svc, cluster, service)
	if err != nil {
		panic(err)
	}

	log.Printf("found service %q (%q)", service, *(s.ServiceArn))

	log.Printf("scaling service %q from %d to %d", service, *(s.DesiredCount), count)

	if _, err := svc.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(cluster),
		Service:      aws.String(service),
		DesiredCount: aws.Int64(count),
	}); err != nil {
		panic(err)
	}

	log.Printf("applied scaling settings")

	if err := pollUntilScaled(svc, cluster, service, count, printEvent); err != nil {
		panic(err)
	}

	log.Printf("deployed successfully")
}
