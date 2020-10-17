# Setup Go

Los pasos para crear el cliente y el servidor en go son los mismos. Describimos los pasos para crear el servidor. Vamos a crear un módulo, para así poder trabajar en una ruta fuera del `GOPATH`. Creamos un directorio productinfo\service, y dentro hacemos:

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

Necesitamos instalar las librerías de gRPC y protobuf para go:

- gRPC library:

```ps
go get -u google.golang.org/grpc
```

- protoc plug-in:

```ps
go get -u github.com/golang/protobuf/protoc-gen-go
```

A partir del protofile crearemos los stubs. Los stubs traducen en forma de structs y funciones los mensajes y servicios del protobuf. Creamos un directorio productinfo\service\ecommerce, y construimos dentro el proto file `product_info.proto`:

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

Generamos los stubs a partir del proto file:

```ps
protoc -I ecommerce ecommerce/product_info.proto --go_out=plugins=grpc:C:\Users\Eugenio\Downloads\gRPC\productinfo\service\ecommerce
```

`protoc` es el compilador de protobuf que habremos instalado previamente. Nos permite generar stubs en diferentes lenguajes. Aquí especificamos go `go_out`. Si no hubieramos indicado la ruta del protofile lo habría tomado del directorio raiz. En nuestro caso hemos especificado otra ruta con `-I`.

Usaremos como punto de partida los stubs para añadir la lógica de negocio. Una vez añadida podemos compilar el programa:

```ps
go build -i -v -o bin/server.exe
```

# Setup Java

Usamos maven. Tenemos que incluir tres dependencias y el plugin de protobuf. Las dependencias son:

```xml
<dependencies>
    <dependency>
      <groupId>junit</groupId>
      <artifactId>junit</artifactId>
      <version>4.11</version>
      <scope>test</scope>
    </dependency>
    <dependency>
       <groupId>io.grpc</groupId>
       <artifactId>grpc-netty</artifactId>
       <version>${io.grpc.version}</version>
   </dependency>
	<dependency>
	  <groupId>io.grpc</groupId>
	  <artifactId>grpc-protobuf</artifactId>
	  <version>${io.grpc.version}</version>
	</dependency>
	<dependency>
	  <groupId>io.grpc</groupId>
	  <artifactId>grpc-stub</artifactId>
	  <version>${io.grpc.version}</version>
	</dependency>
```

El plugin que compila los protofiles:

```xml
<plugin>
    <groupId>org.xolstice.maven.plugins</groupId>
    <artifactId>protobuf-maven-plugin</artifactId>
    <version>${protobuf-maven-plugin.version}</version>
    <configuration>
        <protocArtifact>
            com.google.protobuf:protoc:3.12.0:exe:${os.detected.classifier}
        </protocArtifact>
        <pluginId>grpc-java</pluginId>
        <pluginArtifact>
            io.grpc:protoc-gen-grpc-java:1.32.1:exe:${os.detected.classifier}
        </pluginArtifact>
        <protoSourceRoot>
            ${basedir}/src/main/resources/proto
        </protoSourceRoot>
    </configuration>
    <executions>
        <execution>
            <goals>
                <goal>compile</goal>
                <goal>compile-custom</goal>
            </goals>
        </execution>
    </executions>
</plugin>
```

Podemos definir un directorio donde buscar los prototfiles con `<protoSourceRoot>`. Si no lo especificamos los buscara de `/src/main/proto`. Este otro plugin sirve para detectar en que entorno estamos - windows, mac, ...:

```xml
<plugin>
    <groupId>kr.motd.maven</groupId>
    <artifactId>os-maven-plugin</artifactId>
    <version>${os-maven-plugin.version}</version>
    <executions>
    <execution>
        <phase>initialize</phase>
        <goals>
        <goal>detect</goal>
        </goals>
    </execution>
    </executions>
</plugin>
```