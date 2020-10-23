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

