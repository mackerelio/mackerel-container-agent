package provider

type Type string

const (
	ECS        Type = "ecs"
	Fargate    Type = "fargate"
	Kubernetes Type = "kubernetes"
)
