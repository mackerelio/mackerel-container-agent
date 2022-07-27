package ecs

type provider string

// Supported providers
const (
	ecsProvider     provider = "ecs"
	fargateProvider provider = "fargate"
	// experimental
	externalProvider provider = "external"
)
