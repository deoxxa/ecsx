// ecsx is a set of what I consider to be essential utilities for working with
// Amazon's Elastic Container Service.
//
// Included currently are four commands: `survey`, `update-service-environment`,
// `update-service-image`, and `scale-service`. If you've worked with ECS, these
// should be pretty self explanatory with the exception of `survey`. It will
// iterate through any available ECS clusters, printing some information about
// each service, its containers, their configurations, and any load balancers
// connected.
//
// Usage
//
// See `ecsx --help-long` for the most up-to-date version of this.
//
// ```
// usage: ecsx [<flags>] <command> [<args> ...]
//
// Amazon ECS easy mode
//
// Flags:
//   --help                  Show context-sensitive help (also try --help-long and --help-man).
//   --aws-timeout=1m        Timeout for applying changes.
//   --aws-poll-interval=5s  Interval at which to poll AWS during changes.
//
// Commands:
//   help [<command>...]
//     Show help.
//
//   survey
//     Survey ECS resources and display a summary.
//
//   update-service-environment --cluster=CLUSTER --service=SERVICE --container=CONTAINER [<flags>]
//     Update environment variable(s) for a service.
//
//     --cluster=CLUSTER        ECS cluster name
//     --service=SERVICE        ECS service name
//     --container=CONTAINER    ECS task container
//     --variable=VARIABLE ...  Environment variable to change in KEY=VALUE form
//
//   update-service-image --cluster=CLUSTER --service=SERVICE --container=CONTAINER --image=IMAGE --tag=TAG
//     Update Docker image in use for a service.
//
//     --cluster=CLUSTER      ECS cluster name
//     --service=SERVICE      ECS service name
//     --container=CONTAINER  ECS task container
//     --image=IMAGE          Docker image URL
//     --tag=TAG              Docker image tag
//
//   scale-service --cluster=CLUSTER --service=SERVICE --count=COUNT
//     Scale an ECS service to a specific number.
//
//     --cluster=CLUSTER  ECS cluster name
//     --service=SERVICE  ECS service name
//     --count=COUNT      Desired number of instances
// ```
//
// License
//
// 3-clause BSD. A copy is included with the source.
package main
