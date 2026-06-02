// Command server boots the Gjallarhorn notification service: the kit's REST
// gateway (OpenAPI 3.1) plus a gRPC server exposing GjallarhornService.
package main

import (
	"github.com/fromforgesoftware/go-kit/app"
	"github.com/fromforgesoftware/go-kit/openapi"
	"github.com/fromforgesoftware/go-kit/persistence/gormdb/gormpg"
	kitgrpc "github.com/fromforgesoftware/go-kit/transport/grpc"

	"github.com/fromforgesoftware/gjallarhorn/internal"
)

func main() {
	app.Run(
		app.WithName("gjallarhorn"),
		app.WithVersion(internal.Version),
		app.WithOpenAPI(
			openapi.SpecTitle("Gjallarhorn"),
			openapi.SpecVersion(internal.Version),
			openapi.SpecDescription("Forge multi-channel notification delivery service."),
		),
		gormpg.FxModule(),
		kitgrpc.FxModule(),
		internal.FxModule(),
	)
}
