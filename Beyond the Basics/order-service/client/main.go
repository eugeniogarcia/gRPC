package main

import (
	"context"
	"fmt"
	pb "interceptors/cliente/ecommerce"
	interceptors "interceptors/cliente/interceptors"
	ns "interceptors/cliente/nameservice"

	"io"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc/resolver"

	"github.com/golang/protobuf/ptypes/wrappers"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	address      = "localhost:50051"
	usarDeadline = true
)

//******************************************
//Demuestra el balanceo de carga de cliente
//******************************************

func balanceoCargaPickFirst() {
	//******************************************
	//Demuestra el balanceo de carga de cliente
	//******************************************
	pickfirstConn, errlb := grpc.Dial(
		fmt.Sprintf("%s:///%s", ns.ExampleScheme, ns.ExampleServiceName), // "example:///lb.example.grpc.io"
		// grpc.WithBalancerName("pick_first"), // "pick_first" is the default, so this DialOption is not necessary.
		grpc.WithInsecure(),
	)

	if errlb != nil {
		log.Fatalf("did not connect: %v", errlb)
	}
	defer pickfirstConn.Close()

	log.Println("==== Calling helloworld.Greeter/SayHello with pick_first ====")
	makeRPCs(pickfirstConn, 10)
}

func balanceoCargaRoundrobin() {
	//******************************************
	//Demuestra el balanceo de carga de cliente
	//******************************************
	// Make another ClientConn with round_robin policy.
	roundrobinConn, errlb := grpc.Dial(
		fmt.Sprintf("%s:///%s", ns.ExampleScheme, ns.ExampleServiceName), // // "example:///lb.example.grpc.io"
		grpc.WithBalancerName("round_robin"),                             // This sets the initial balancing policy.
		grpc.WithInsecure(),
	)
	if errlb != nil {
		log.Fatalf("did not connect: %v", errlb)
	}
	defer roundrobinConn.Close()

	log.Println("==== Calling helloworld.Greeter/SayHello with round_robin ====")
	makeRPCs(roundrobinConn, 10)
}

func makeRPCs(cc *grpc.ClientConn, n int) {
	hwc := pb.NewOrderManagementClient(cc)
	for i := 0; i < n; i++ {
		order := pb.Order{Id: strconv.Itoa(i + 10), Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}

		callUnaryOrder(hwc, order)
	}
}

func callUnaryOrder(c pb.OrderManagementClient, message pb.Order) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.AddOrder(ctx, &message)
	if err != nil {
		got := status.Code(err)
		log.Printf("Error Occured -> addOrder : , %v:", got)
		log.Fatalf("could not greet: %v", err)
	} else {
		log.Print("AddOrder Response -> ", r.Value)
	}
}

//******************************************
//Llamadas con interceptor
//******************************************

func usaInterceptors() {
	// Conexion con el servidor. Configura interceptors
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptors.OrderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(interceptors.ClientStreamInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)

	//Contexto que vamos a usar en las llamadas
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	unitarioRPC(ctx, client)

	streamServidorRPC(ctx, client)

	streamClienteRPC(ctx, client)

	streamBidireccionalRPC(ctx, client)
}

func unitarioRPC(ctx context.Context, client pb.OrderManagementClient) {

	order1 := pb.Order{Id: "101", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}

	res, addErr := client.AddOrder(ctx, &order1)
	if addErr != nil {
		got := status.Code(addErr)
		log.Printf("Error Occured -> addOrder : , %v:", got)
	} else {
		log.Print("AddOrder Response -> ", res.Value)
	}

	// Get Order
	retrievedOrder, err := client.GetOrder(ctx, &wrapper.StringValue{Value: "106"})
	if err != nil {
		got := status.Code(err)
		log.Printf("Error Occured -> addOrder : , %v:", got)
	} else {
		log.Print("GetOrder Response -> : ", retrievedOrder)
	}

}

func streamServidorRPC(ctx context.Context, client pb.OrderManagementClient) {
	// Search Order : Server streaming scenario
	searchStream, _ := client.SearchOrders(ctx, &wrapper.StringValue{Value: "Google"})
	for {
		searchOrder, err := searchStream.Recv()
		if err == io.EOF {
			log.Print("EOF")
			break
		}

		if err == nil {
			log.Print("Search Result : ", searchOrder)
		}
	}
}

func streamClienteRPC(ctx context.Context, client pb.OrderManagementClient) {

	updOrder1 := pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Google Pixel Book"}, Destination: "Mountain View, CA", Price: 1100.00}

	updOrder2 := pb.Order{Id: "103", Items: []string{"Apple Watch S4", "Mac Book Pro", "iPad Pro"}, Destination: "San Jose, CA", Price: 2800.00}

	updOrder3 := pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub", "iPad Mini"}, Destination: "Mountain View, CA", Price: 2200.00}

	updateStream, err := client.UpdateOrders(ctx)

	if err != nil {
		log.Fatalf("%v.UpdateOrders(_) = _, %v", client, err)
	}

	//Enviamos tres mensajes por el stream
	// Updating order 1
	if err := updateStream.Send(&updOrder1); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder1, err)
	}

	// Updating order 2
	if err := updateStream.Send(&updOrder2); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder2, err)
	}

	// Updating order 3
	if err := updateStream.Send(&updOrder3); err != nil {
		log.Fatalf("%v.Send(%v) = %v", updateStream, updOrder3, err)
	}

	//Indicamos que este era el último mensaje del stream
	updateRes, err := updateStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", updateStream, err, nil)
	}
	log.Printf("Update Orders Res : %s", updateRes)
}

func asncClientBidirectionalRPC(streamProcOrder pb.OrderManagement_ProcessOrdersClient, c chan bool) {
	for {
		combinedShipment, errProcOrder := streamProcOrder.Recv()
		if errProcOrder == io.EOF {
			break
		}
		log.Printf("Combined shipment : ", combinedShipment.OrdersList)
	}
	c <- true
}

func streamBidireccionalRPC(ctx context.Context, client pb.OrderManagementClient) {
	// =========================================
	// Process Order : Bi-di streaming scenario
	streamProcOrder, err := client.ProcessOrders(ctx)
	if err != nil {
		log.Fatalf("%v.ProcessOrders(_) = _, %v", client, err)
	}

	//Enviamos tres mensajes por el stream
	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "102"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "102", err)
	}

	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "103"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "103", err)
	}

	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "104"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "104", err)
	}

	//Nos ponemos a escuchar por el stream del servidor
	channel := make(chan bool)
	go asncClientBidirectionalRPC(streamProcOrder, channel)
	time.Sleep(time.Millisecond * 1000)

	//Enviamos un mensaje más al servidor
	if err := streamProcOrder.Send(&wrapper.StringValue{Value: "101"}); err != nil {
		log.Fatalf("%v.Send(%v) = %v", client, "101", err)
	}
	//Indicamos al servidor que no vamos a enviar más mensajes por el stream. Es el "half-close" de la comunicación
	if err := streamProcOrder.CloseSend(); err != nil {
		log.Fatal(err)
	}
	//Esperamos hasta que obtengamos una respuesta en el canal. Esto nos permite sincronizarnos con la go-rutina que esta escuchando por el stream del servidor, de forma que el cliente termine solo cuando ya no se vayan a recibir más mensajes desde el servidor
	<-channel
}

//******************************************
//Llamadas con deadline
//Llamadas con compresion
//demuestra gestión de errores
//******************************************

func llamaConDeadline(client pb.OrderManagementClient, duracion time.Duration, compresion bool) {
	clientDeadline := time.Now().Add(time.Duration(duracion * time.Second))
	ctx, cancel := context.WithDeadline(context.Background(), clientDeadline)
	defer cancel()

	order := pb.Order{Id: "101", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}

	var res *wrappers.StringValue
	var addErr error
	if compresion {
		res, addErr = client.AddOrder(ctx, &order, grpc.UseCompressor(gzip.Name))
	} else {
		res, addErr = client.AddOrder(ctx, &order)
	}

	//Si devuelve un error...
	if addErr != nil {
		//Obtenemos el código de error
		errorCode := status.Code(addErr)
		if errorCode == codes.InvalidArgument {
			log.Printf("Invalid Argument Error : %s", errorCode)

			//Obtenemos el detalle asociado al error
			errorStatus := status.Convert(addErr)
			for _, d := range errorStatus.Details() {
				//Comprueba el tipo informado en el detalle. Esperamos encontrar un puntero a epb.BadRequest_FieldViolation
				switch info := d.(type) {
				case *epb.BadRequest_FieldViolation:
					log.Printf("Request Field Invalid: %s", info)
				default:
					log.Printf("Unexpected error type: %s", info)
				}
			}
		} else {
			log.Printf("Unhandled error : %s ", errorCode)
		}
	} else {
		log.Print("AddOrder Response -> ", res.Value)
	}

}

//******************************************
//Llamadas con metadatos
//******************************************

func llamaConMetadatos(client pb.OrderManagementClient) {
	//Primera forma de crear metadatos. Añadiendo duplas
	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"kn", "vn",
	)

	//Crea el contexto con los metadatos. Machacaría cualquier metadato que se hubiera añadido previamente al contexto
	mdCtx := metadata.NewOutgoingContext(context.Background(), md)

	//Segunda forma de crear metadatos. Añadiendo metadatos a un contexto ya existente
	ctxA := metadata.AppendToOutgoingContext(mdCtx, "k1", "v1", "k1", "v2", "k2", "v3")

	//Hacemos la llamada
	// Variables en las que vamos a guardar la cabecera y metadatos de la respuesta. Ambas son del tipo MD
	var header, trailer metadata.MD

	order := pb.Order{Id: "1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}

	res, _ := client.AddOrder(ctxA, &order, grpc.Header(&header), grpc.Trailer(&trailer))

	log.Print("AddOrder Response -> ", res.Value)

	// Obtenemos el valor de las cabeceras. Los metadatos son transportados como cabeceras custom
	if t, ok := header["timestamp"]; ok {
		log.Printf("timestamp from header:\n")
		for i, e := range t {
			fmt.Printf(" %d. %s\n", i, e)
		}
	} else {
		log.Fatal("timestamp expected but doesn't exist in header")
	}

}

//******************************************

func main() {

	balanceoCargaPickFirst()

	balanceoCargaRoundrobin()

	usaInterceptors()

	// Setting up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderManagementClient(conn)

	llamaConDeadline(client, 2, false)

	llamaConDeadline(client, 2, true)

	llamaConMetadatos(client)

}

func init() {
	resolver.Register(&ns.ExampleResolverBuilder{})
}
