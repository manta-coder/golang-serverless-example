package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/caarlos0/env/v6"
	"github.com/childrenofukiyo/odin/pkg/auth"
	"github.com/childrenofukiyo/odin/pkg/controller"
	"github.com/childrenofukiyo/odin/pkg/odin"
	"github.com/childrenofukiyo/odin/pkg/service"
	"github.com/childrenofukiyo/odin/pkg/store"
	"time"
)

var echoLambda *echoadapter.EchoLambdaV2

func init() {
	var config odin.Config
	if err := env.Parse(&config); err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	server := odin.MustServer(config)

	userStore := store.NewUserStore(server.Logger, server.DB)
	challengeStore := store.NewChallengeStore(server.Logger, server.DB)

	ted := time.Duration(config.AuthTokenExpiryDurationSeconds) * time.Second

	userService := service.NewUserService(server.Logger, userStore)
	authentication := auth.NewService(config.AuthSecret, ted)
	authService := service.NewAuthService(server.Logger, authentication, challengeStore, userService)

	authenticator := controller.NewAuthenticator(config.AuthSecret)

	group := server.Echo.Group("/auth")
	controller.NewAuthController(group, server.Logger, authService, authenticator)

	echoLambda = echoadapter.NewV2(server.Echo)
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(handler)
}
