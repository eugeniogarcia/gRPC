# Setup Go

- Crear un módulo de go

Creamos un directorio productinfo\service, y dentro hacemos:

```ps
go mod init productinfo/service
```

Esto nos crea un archivo llamado `go.mod`. Incluimos en el arvhivo los imports que vamos a necesitar en este módulo:

```go
module productinfo/service

go 1.15

require (
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.4.3
	github.com/google/uuid v1.1.2
	google.golang.org/grpc v1.33.0
	google.golang.org/protobuf v1.25.0

)
```

Necesitamos algunos componentes de go:

- gRPC library:

```ps
go get -u google.golang.org/grpc
```

- protoc plug-in:

```ps
go get -u github.com/golang/protobuf/protoc-gen-go
```

- Creamos un directorio productinfo\service\ecommerce, y construimos dentro el proto file `product_info.proto`:

```proto
syntax = "proto3";

package ecommerce;

service ProductInfo {
    rpc addProduct(Product) returns (ProductID);
    rpc getProduct(ProductID) returns (Product);
}

message Product {
    string id = 1;
    string name = 2;
    string description = 3;
    float price = 4;
}

message ProductID {
    string value = 1;
}
```

- Generamos los stubs a partir del proto file:

```ps
protoc -I ecommerce ecommerce/product_info.proto --go_out=plugins=grpc:C:\Users\Eugenio\Downloads\gRPC\productinfo\service\ecommerce
```

Una vez construida la lógica de negocio podemos compilar el programa

```ps
go build -i -v -o bin/server.exe
```




```ps
java -jar build/libs/server.jar
```