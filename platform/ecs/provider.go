package ecs

type provider string

// Supported providers
const (
	ecsProvider        provider = "ecs"
	fargateProvider    provider = "fargate"
	ecsManagedProvider provider = "ecs-managed"
	// experimental
	ecsAnywhereProvider provider = "ecs-anywhere"
)
