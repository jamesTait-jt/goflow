package infrastructure

var (
	RedisContainerName = "goflow-redis-container"

	// RedisPort is the port that Redis will listen on inside the container
	RedisPort int32 = 6379

	GRPCServerContainerName = "goflow-grpc-container"

	// GRPCContainerPort is the port that the gRPC server will listen on inside the container
	GRPCContainerPort int32 = 50051

	// WorkerpoolHandlersLocation is the location of the handlers on the pluginBuilder and workerpool containers
	WorkerpoolHandlersLocation = "/app/handlers"
	WorkerpoolContainerName    = "workerpool-container"
)
