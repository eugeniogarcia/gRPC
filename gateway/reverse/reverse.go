package main

import (
	"context"
	"log"
	"net/http"

	gw "gz.com/gateway/ecommerce"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

var (
	grpcServerEndpoint = "localhost:50051"
)

func main() {
	//Creamos el contexto, sin timeouts, etc., con los valores por defecto
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	//Registra el servidor RPC y nos crea un mux
	err := gw.RegisterProductInfoHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)

	if err != nil {
		log.Fatalf("Fail to register gRPC service endpoint: %v", err)
		return
	}

	//Arranca el servidor http
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Could not setup HTTP endpoint: %v", err)
	}
}
