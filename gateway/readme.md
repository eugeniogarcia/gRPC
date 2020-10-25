# Stubs

## Setup

Necesitamos instalar estas dependencias:

```ps
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

go get -u github.com/golang/protobuf/protoc-gen-go
```

## Generar Stubs

Necesitamos importar en el protofile las anotaciones que nos van a permitir configurar como deben generarse los endpoints REST:

```proto
import "google/api/annotations.proto";
```

Las anotaciones nos permitirán extender la definición de nuestro protobuffer para indicar como generar los REST endpoints. En este caso estamos definiendo dos endpoints, un POST y un GET:

```proto
service ProductInfo {
    rpc addProduct(Product) returns (google.protobuf.StringValue) {
        option (google.api.http) = {
            post: "/v1/product"
            body: "*"
        };
    }
    rpc getProduct(google.protobuf.StringValue) returns (Product) {
         option (google.api.http) = {
             get:"/v1/product/{value}"
         };
    }
}
```

Para generar los stubs del servicio RPC usamos el compilador de la forma habitual:

```ps
protoc --go_out=plugins=grpc:..\ecommerce\ product_info.proto
```

Para generar el reverse proxy:

```ps
protoc --grpc-gateway_out=logtostderr=true:..\reverse\ecommerce\ product_info.proto

protoc --grpc-gateway_out=logtostderr=true:..\backend\ecommerce\ product_info.proto
```

Podemos generar el swagger:

```ps
protoc --swagger_out=logtostderr=true:. product_info.proto
```

Con esto tenemos todos los stubs que necesitamos generados.

# Reverse Proxy

Para crear el reverse proxy, creamos primero el contexto:

```go
ctx := context.Background()
ctx, cancel := context.WithCancel(ctx)
defer cancel()
```

Creamos un servidor RPC que exponga los end-points como REST. Este componente es parte de `"github.com/grpc-ecosystem/grpc-gateway/runtime"`:

```go
mux := runtime.NewServeMux()
opts := []grpc.DialOption{grpc.WithInsecure()}
```

Registramos el el backend:

```go
//Registra el servidor RPC y nos crea un mux
err := gw.RegisterProductInfoHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
```

Finalmente abrimos el servidor REST:

```go
if err := http.ListenAndServe(":8081", mux); err != nil {
    log.Fatalf("Could not setup HTTP endpoint: %v", err)
}
```

Con esto hemos abierto un servidor Http que expone los servicios REST y deriva la petición como gRPC al backend.

El backend es un servidor normal. Hemos incluido un proyecto Postman con los endpoints