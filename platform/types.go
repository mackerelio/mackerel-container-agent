package platform

// Type represents supported platform types
type Type string

// Supported platform types
const (
	ECS        Type = "ecs"
	ECSAwsvpc  Type = "ecs_awsvpc"
	Fargate    Type = "fargate"
	Kubernetes Type = "kubernetes"
)
