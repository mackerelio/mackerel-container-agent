package ecs

type provider string

// Supported providers
const (
	ecsProvider     provider = "ecs"
	fargateProvider provider = "fargate"
	// experimental
	ecsAnywhereProvider provider = "ecs-anywhere"
)
