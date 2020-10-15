- Crear un módulo de go

```ps
go mod init productinfo/service
```

- gRPC library:

```ps
go get -u google.golang.org/grpc
```

- protoc plug-in:

```ps
go get -u github.com/golang/protobuf/protoc-gen-go
```

- proto file in ecommerce/product_info.proto:

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

- Generar los stubs:

```ps
protoc -I ecommerce ecommerce/product_info.proto --go_out=plugins=grpc:C:\Users\Eugenio\Downloads\gRPC\productinfo\service\ecommerce
```

```ps
go build -i -v -o bin/server
```

```ps
go build -i -v -o bin/client
```

```ps
java -jar build/libs/server.jar
```