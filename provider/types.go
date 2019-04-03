package provider

// Type represents supported providers
type Type string

// Supported providers
const (
	ECS        Type = "ecs"
	Fargate    Type = "fargate"
	Kubernetes Type = "kubernetes"
)
