package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("ecsx", "Amazon ECS easy mode")

	awsTimeout      = app.Flag("aws-timeout", "Timeout for applying changes.").Default("1m").Duration()
	awsPollInterval = app.Flag("aws-poll-interval", "Interval at which to poll AWS during changes.").Default("5s").Duration()

	survey = app.Command("survey", "Survey ECS resources and display a summary.")

	updateServiceEnvironment          = app.Command("update-service-environment", "Update environment variable(s) for a service.")
	updateServiceEnvironmentCluster   = updateServiceEnvironment.Flag("cluster", "ECS cluster name").Required().String()
	updateServiceEnvironmentService   = updateServiceEnvironment.Flag("service", "ECS service name").Required().String()
	updateServiceEnvironmentContainer = updateServiceEnvironment.Flag("container", "ECS task container").Required().String()
	updateServiceEnvironmentVariable  = updateServiceEnvironment.Flag("variable", "Environment variable to change in KEY=VALUE form").Strings()

	updateServiceImage          = app.Command("update-service-image", "Update Docker image in use for a service.")
	updateServiceImageCluster   = updateServiceImage.Flag("cluster", "ECS cluster name").Required().String()
	updateServiceImageService   = updateServiceImage.Flag("service", "ECS service name").Required().String()
	updateServiceImageContainer = updateServiceImage.Flag("container", "ECS task container").Required().String()
	updateServiceImageImage     = updateServiceImage.Flag("image", "Docker image URL").Required().String()
	updateServiceImageTag       = updateServiceImage.Flag("tag", "Docker image tag").Required().String()

	scaleService        = app.Command("scale-service", "Scale an ECS service to a specific number.")
	scaleServiceCluster = scaleService.Flag("cluster", "ECS cluster name").Required().String()
	scaleServiceService = scaleService.Flag("service", "ECS service name").Required().String()
	scaleServiceCount   = scaleService.Flag("count", "Desired number of instances").Required().Int64()
)

func main() {
	c := kingpin.MustParse(app.Parse(os.Args[1:]))

	awsSession := session.New()

	switch c {
	case survey.FullCommand():
		surveyCommand(awsSession)
	case updateServiceEnvironment.FullCommand():
		updateServiceEnvironmentCommand(awsSession, *updateServiceEnvironmentCluster, *updateServiceEnvironmentService, *updateServiceEnvironmentContainer, *updateServiceEnvironmentVariable)
	case updateServiceImage.FullCommand():
		updateServiceImageCommand(awsSession, *updateServiceImageCluster, *updateServiceImageService, *updateServiceImageContainer, *updateServiceImageImage, *updateServiceImageTag)
	case scaleService.FullCommand():
		scaleServiceCommand(awsSession, *scaleServiceCluster, *scaleServiceService, *scaleServiceCount)
	default:
		log.Fatalf("unrecognised command %q", c)
	}
}

func pollUntil(svc *ecs.ECS, cluster, service string, ready func(s *ecs.Service) bool, fn func(ev *ecs.ServiceEvent)) error {
	lastSeen := time.Now().Add(-1 * time.Minute)
	deadline := time.Now().Add(*awsTimeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout")
		}

		s, err := getService(svc, cluster, service)
		if err != nil {
			return err
		}

		for i := len(s.Events) - 1; i >= 0; i-- {
			event := s.Events[i]

			if event.CreatedAt.After(lastSeen) {
				fn(event)
				lastSeen = *event.CreatedAt
			}
		}

		if ready(s) {
			return nil
		}

		time.Sleep(*awsPollInterval)
	}
}

func pollUntilScaled(svc *ecs.ECS, cluster, service string, count int64, fn func(ev *ecs.ServiceEvent)) error {
	return pollUntil(svc, cluster, service, func(s *ecs.Service) bool {
		return *(s.RunningCount) == count
	}, fn)
}

func pollUntilTaskDeployed(svc *ecs.ECS, cluster, service, task string, fn func(ev *ecs.ServiceEvent)) error {
	return pollUntil(svc, cluster, service, func(s *ecs.Service) bool {
		return len(s.Deployments) == 1 && *(s.Deployments[0].TaskDefinition) == task
	}, fn)
}

func printEvent(ev *ecs.ServiceEvent) {
	fmt.Printf("> %s %s\n", ev.CreatedAt.Format("2006/01/02 15:04:05"), *(ev.Message))
}

func getService(svc *ecs.ECS, cluster, service string) (*ecs.Service, error) {
	l, err := svc.DescribeServices(&ecs.DescribeServicesInput{
		Cluster:  aws.String(cluster),
		Services: aws.StringSlice([]string{service}),
	})
	if err != nil {
		panic(err)
	}

	if len(l.Services) != 1 {
		return nil, fmt.Errorf("couldn't find service %q in cluster %q", service, cluster)
	}

	return l.Services[0], nil
}

func findContainerByName(d *ecs.TaskDefinition, n string) *ecs.ContainerDefinition {
	for _, c := range d.ContainerDefinitions {
		if c.Name != nil && *c.Name == n {
			return c
		}
	}

	return nil
}

func findVariableByName(c *ecs.ContainerDefinition, n string) *ecs.KeyValuePair {
	for _, p := range c.Environment {
		if p.Name != nil && *p.Name == n {
			return p
		}
	}

	return nil
}
