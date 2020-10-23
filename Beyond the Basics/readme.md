# Interceptors

Hay dos tipos de interceptors:

- cliente
- servidor

Los interceptors tambien se pueden clasificar como:

- Unitarios. Interceptan llamadas que siguen el patrón de comunicación Unitario
- Stream. Interceptan llamadas que siguen cualquiera de los tres patrones en los que se usan streams

## Servidor

### Unitario

```go
func orderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Pre-processing logic
	// Gets info about the current RPC call by examining the args passed in
	log.Println("======= [Interceptor Unitario en el Servidor] ", info.FullMethod)
	log.Printf(" Preprocesa el mensaje : %s", req)

	// Invoking the handler to complete the normal execution of a unary RPC.
	m, err := handler(ctx, req)

	// Post processing logic
	log.Printf(" Postprocesa la respuesta : %s", m)
	return m, err
}
```

- Los argumentos son:
    - mensaje de entrada: `req`. Esta definido como un tipo generico `interface{}`
    - El handler. Hace referencia a la llamada RPC
    - Información de la llamada - método, recurso, ... `*grpc.UnaryServerInfo`
    - Contexto gRPC
- Podemos manipular el mesaje antes de hacer la llamada RPC
- Podemos manipular el mensaje de respuesta de la llamada RPC
- Podemos decidir si finalmente hacer o no la llamada RPC

### Stream

```go
//Interceptador de streams
func orderServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Pre-processing
	log.Println("====== [Interceptador de Streams en el servidor] ", info.FullMethod)

	//Procesa el stream con el wrapper. El wrapper se encarga de recivir y enviar los mensajes. Este método no retorna ningun mensaje
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}
	return err
}
```

- Los argumentos son:
    - El handler. Hace referencia a la llamada RPC. En este interceptador es de tipo `grpc.StreamHandler`
    - Información de la llamada - método, recurso, ... `*grpc.StreamServerInfo`
    - No tenemos acceso al mensaje de entrada sino al stream. `grpc.ServerStream`. El stream se tiene que manipular en un wrapper, y será el wrapper lo que pasemos como argumento al handler:

    ```go
    err := handler(srv, newWrappedStream(ss))
    ```

El wrapper nos permitirá acceder a los mensajes que se envia o se reciben. Para ello el wrapper se define como:

```go
type wrappedStream struct {
	grpc.ServerStream
}
```

Este tipo `grpc.ServerStream` implementa un interface que tiene dos métodos, uno que intercepta los mensajes recibidos por el stream:

```go
//Método que procesa los mensajes recibidos por el stream
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del servidor] Se recivio un mensaje (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}
```

Y otro que intercepta los mensajes enviados por el stream:

```go
//Método que procesa los mensajes enviados por el stream
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del servidor] Se envia un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}
```

## Cliente

Es análogo a lo que hemos visto para el servidor

### Unitario

```go
func orderUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Pre-processor phase
	log.Println("======= [Interceptor Unitario en el Cliente] ", method)
	log.Printf(" Preprocesa el mensaje : %s", req)

	// Invoking the remote method
	err := invoker(ctx, method, req, reply, cc, opts...)

	log.Printf(" Postprocesa la respuesta : %s", reply)

	return err
}
```

- Los argumentos son:
    - mensaje de entrada: `req`
    - mensaje de respuesta: `reply`. En el caso del servidor se definía como la respuesta de la función. En el caso del cliente es un argumento de la función
    - El invoker. Juega el mismo papel que jugaba el handler en el servidor, hace referencia a la llamada RPC. `grpc.UnaryInvoker`
    - Información de la llamada: method, cliente,...
    - Contexto gRPC
- Podemos manipular el mesaje antes de hacer la llamada RPC
- Podemos manipular el mensaje de respuesta de la llamada RPC
- Podemos decidir si finalmente hacer o no la llamada RPC

### Stream

```go
func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	log.Println("======= [Interceptador de Streams en el cliente	] ", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}
```

Es muy similar al caso del servidor. También se usa un wrapper para procesar los mensajes enviados y recibidos:

```go
type wrappedStream struct {
	grpc.ClientStream
}
```

`grpc.ClientStream` implementa un interface que incluye un método para recivir mensajes:

```go
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del cliente] Recive un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}
```

y otro que intercepta los enviados:

```go
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del cliente] Envia un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}
```

## Configuración

Veamos como se tiene que configurar gRPC para que se usen los interceptadores.

### Servidor

```go
s := grpc.NewServer(
		grpc.UnaryInterceptor(orderUnaryServerInterceptor),
		grpc.StreamInterceptor(orderServerStreamInterceptor))
    
    pb.RegisterOrderManagementServer(s, &server{})
```

### Cliente

```go
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor))

    if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
    
    defer conn.Close()
    
    client := pb.NewOrderManagementClient(conn)
```

# Deadlines

Un deadline permite especificar una duración máxima a una petición, pero de tal forma que aplica a todas las llamadas RPC 
Aplica al cliente. Al crear el contexto

```go
	if usarDeadline {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	} else {
		clientDeadline := time.Now().Add(time.Duration(20 * time.Second))
		ctx, cancel = context.WithDeadline(context.Background(), clientDeadline)
    }
```

En cualquier momento podemos comprobar si se ha superado el deadline. Por ejemplo, en el servidor podríamos hacer:

```go
if ctx.Err() == context.DeadlineExceeded {
		log.Printf("RPC has reached deadline exceeded state : %s", ctx.Err())
		return nil, ctx.Err()
    }
```

## Cancelación

Si deseamos cancelar una ejecución bastaría con llamar al método `cancel`:

```go
	if usarDeadline {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	} else {
		clientDeadline := time.Now().Add(time.Duration(20 * time.Second))
		ctx, cancel = context.WithDeadline(context.Background(), clientDeadline)
    }

....

cancel()
```

# Gestión de Errores

## Crear un error

Podemos crear un error personalizado:

```go
//Crea un error
errorStatus := status.New(codes.InvalidArgument, "Invalid information received")
```

El error puede tener asociada información de detalle:

```go
//Podemos añadir detalles al error
ds, err := errorStatus.WithDetails(
    &epb.BadRequest_FieldViolation{
        Field:       "ID",
        Description: fmt.Sprintf("Order ID received is not valid %s : %s", orderReq.Id, orderReq.Description),
    },
```

En la información de detalle hemos incluido un mensaje. En el paquete `epb "google.golang.org/genproto/googleapis/rpc/errdetails"` tenemos varios [mensajes tipo](https://godoc.org/google.golang.org/genproto/googleapis/rpc/errdetails).

## Procesar un error

Al hacer una llamada obtenemos tambien el posible error:

```go
res, addOrderError := client.AddOrder(ctx, &order1_err)
```

Podemos recuperar el código del error:

```go
//Si devuelve un error...
if addOrderError != nil {
    //Obtenemos el código de error
    errorCode := status.Code(addOrderError)
    if errorCode == codes.InvalidArgument {
        log.Printf("Invalid Argument Error : %s", errorCode)
```

También podemos acceder a los detalles del error:

```go
//Obtenemos el detalle asociado al error
errorStatus := status.Convert(addOrderError)

for _, d := range errorStatus.Details() {
    //Comprueba el tipo informado en el detalle. Esperamos encontrar un puntero a epb.BadRequest_FieldViolation
    switch info := d.(type) {
    case *epb.BadRequest_FieldViolation:
        log.Printf("Request Field Invalid: %s", info)
    default:
        log.Printf("Unexpected error type: %s", info)
    }
}
```

Podemos observar como verificamos el tipo que obtenemos con el detalle, y verificamos si coincide con `*epb.BadRequest_FieldViolation`.

# Multiplexing

## Cliente

Podemos usar el mismo canal/conexión para contactar con diferentes servicios RPC. Por ejemplo, establecemos una conexión:

```go
conn, err := grpc.Dial(address, grpc.WithInsecure())
```

Usamos la conexión para utilizar el servicio RPC:

```go
orderManagementClient := pb.NewOrderManagementClient(conn)

order1 := pb.Order{Id: "101", Items:[]string{"iPhone XS", "Mac Book Pro"}, Destination:"San Jose, CA", Price:2300.00}

res, addErr := orderManagementClient.AddOrder(ctx, &order1)
```

Y la misma conexión para contactar con otro servicio RPC:

```go
helloClient := hwpb.NewGreeterClient(conn)
        
helloResponse, err := helloClient.SayHello(hwcCtx, &hwpb.HelloRequest{Name: "gRPC Up and Running!"})
```

Notese que en el ejemplo anterior usamos dos contextos diferentes - podríamos también haber usado el mismo.


## Servidor

podemos registrar por el mismo canal varios servicios RPC:

```go
grpcServer := grpc.NewServer()

// Register Order Management service on gRPC orderMgtServer
ordermgt_pb.RegisterOrderManagementServer(grpcServer, &orderMgtServer{})
```

Y podríamos registrar un segundo servicio RPC:

```go
// Register Greeter Service on gRPC orderMgtServer
hello_pb.RegisterGreeterServer(grpcServer, &helloServer{})
```

# Metadata

Los metadatos que incluyamos en una llamada viajaran como una cabecera más que se añadirá a las cabeceras estandard. Los metadatos los incluimos en el contextos de la llamada

## Cliente

### Crear metadatos

Los metadatos se especifican en el contexto. Podemos crear un conjunto de metadatos informado parejas de key, value, y luego crear el contexto con estos metadatos:

```go
//Primera forma de crear metadatos. Añadiendo duplas
md := metadata.Pairs(
    "timestamp", time.Now().Format(time.StampNano),
    "kn", "vn",
)
//Crea el contexto con los metadatos. Machacaría cualquier metadato que se hubiera añadido previamente al contexto
mdCtx := metadata.NewOutgoingContext(context.Background(), md)
```

Si por el contrario tenemos un contexto, que a lo mejor ya tiene algún metadato especificado, y queremos añadir más metadatos, podemos actualizar el contexto:

```go
//Segunda forma de crear metadatos. Añadiendo metadatos a un contexto ya existente
ctxA := metadata.AppendToOutgoingContext(mdCtx, "k1", "v1", "k1", "v2", "k2", "v3")
```

Aquí hemos creado un nuevo contexto, `ctxA` a partir de uno ya existente `mdCtx`. Si hacemos la llamada con este contexto viajaran los metadatos.

### Leer metadatos de la respuesta

Supongamos que queremos acceder a la cabecera  y metadatos de la respuesta. Para ello haremos que se guarde la cabecera de la respuesta en una variable:

```go
//Variable donde guardar la cabecera de la respuesta
var header, trailer metadata.MD

// RPC: Add Order
order1_md := pb.Order{Id: "1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
res, _ = client.AddOrder(ctxA, &order1_md, grpc.Header(&header), grpc.Trailer(&trailer))
```

Podemos acceder a las cabeceras:

```go
if t, ok := header["timestamp"]; ok {
    log.Printf("timestamp from header:\n")
    for i, e := range t {
        fmt.Printf(" %d. %s\n", i, e)
    }
} else {
    log.Fatal("timestamp expected but doesn't exist in header")
}
```

## Servidor

### Leer metadatos de la petición

Recuperamos los metadatos accediendo al contexto:

```go
//Obtiene los metadatos del contexto
md, metadataAvailable := metadata.FromIncomingContext(ctx)

//Si no hay metadatos devuelve un error
if !metadataAvailable {
    return nil, status.Errorf(codes.DataLoss, "UnaryEcho: failed to get metadata")
}
```

Podemos obtener el valor del metadato de la misma forma que en el cliente:

```go
//Si hay metadatos, recupera el metadato timestamp, y los detalles asociados
if t, ok := md["timestamp"]; ok {
    fmt.Printf("timestamp from metadata:\n")
    for i, e := range t {
        fmt.Printf("====> Metadata %d. %s\n", i, e)
    }
}
```

### Escribir metadatos en la respuesta

Añadir metadatos en la respuesta:

```go
//Creamos metadatos
header := metadata.New(map[string]string{"location": "San Jose", "timestamp": time.Now().Format(time.StampNano)})

//Los añadimos a la cabecera de respuesta
grpc.SendHeader(ctx, header)
```

### Stream

Podemos escribir los metadatos bien en el trailer - el trailer se envia con el frame que indica el final del stream:

```go
//Añade los metadatos al trailer
defer func() {
    trailer := metadata.Pairs("timestamp", time.Now().Format(time.StampNano))
    stream.SetTrailer(trailer)
}()
```

O podemos añadirlos a la cabecera - de cada mensaje del stream:

```go
//Añade los metadatos al header
header := metadata.New(map[string]string{"location": "MTV", "timestamp": time.Now().Format(time.StampNano)})
stream.SendHeader(header)
```

