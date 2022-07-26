package platform

// Type represents supported platform types
type Type string

// Supported platform types
const (
	ECS          Type = "ecs"
	ECSAwsvpc    Type = "ecs_awsvpc"
	ECSv3        Type = "ecs_v3"
	Fargate      Type = "fargate"
	Kubernetes   Type = "kubernetes"
	EKSOnFargate Type = "eks_fargate"
	None         Type = "none"
	// experimental
	ECSExternal Type = "ecs_external"
)
