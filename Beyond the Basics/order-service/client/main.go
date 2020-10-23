package main

import (
	"context"
	"fmt"
	pb "interceptors/cliente/ecommerce"
	"io"
	"log"
	"time"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	address      = "localhost:50051"
	usarDeadline = true
)

func main() {
	// Setting up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(orderUnaryClientInterceptor),
		grpc.WithStreamInterceptor(clientStreamInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewOrderManagementClient(conn)

	var ctx context.Context
	var cancel context.CancelFunc
	if usarDeadline {
		clientDeadline := time.Now().Add(time.Duration(2 * time.Second))
		ctx, cancel = context.WithDeadline(context.Background(), clientDeadline)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	}
	defer cancel()

	// Add Order
	order1 := pb.Order{Id: "101", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	res, addErr := client.AddOrder(ctx, &order1)
	if addErr != nil {
		got := status.Code(addErr)
		log.Printf("Error Occured -> addOrder : , %v:", got)
	} else {
		log.Print("AddOrder Response -> ", res.Value)
	}

	// Add Order
	// This is an invalid order
	order1_err := pb.Order{Id: "-1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	res, addOrderError := client.AddOrder(ctx, &order1_err)

	//Si devuelve un error...
	if addOrderError != nil {
		//Obtenemos el código de error
		errorCode := status.Code(addOrderError)
		if errorCode == codes.InvalidArgument {
			log.Printf("Invalid Argument Error : %s", errorCode)
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
		} else {
			log.Printf("Unhandled error : %s ", errorCode)
		}
	} else {
		log.Print("AddOrder Response -> ", res.Value)
	}

	//Usa metadatos

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
	// Variables en las que vamos a guardar la cabecera y metadatos de la respuesta
	var header, trailer metadata.MD

	// RPC: Add Order
	order1_md := pb.Order{Id: "1", Items: []string{"iPhone XS", "Mac Book Pro"}, Destination: "San Jose, CA", Price: 2300.00}
	res, _ = client.AddOrder(ctxA, &order1_md, grpc.Header(&header), grpc.Trailer(&trailer))

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

	if l, ok := header["location"]; ok {
		log.Printf("location from header:\n")
		for i, e := range l {
			fmt.Printf(" %d. %s\n", i, e)
		}
	} else {
		log.Fatal("location expected but doesn't exist in header")
	}

	// Get Order
	retrievedOrder, err := client.GetOrder(ctx, &wrapper.StringValue{Value: "106"})
	log.Print("GetOrder Response -> : ", retrievedOrder)

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

	// =========================================
	// Update Orders : Client streaming scenario
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

func orderUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	// Pre-processor phase
	log.Println("======= [Interceptor Unitario en el Cliente] ", method)
	log.Printf(" Preprocesa el mensaje : %s", req)

	// Invoking the remote method
	err := invoker(ctx, method, req, reply, cc, opts...)

	if err == nil {
		log.Printf(" Postprocesa la respuesta : %s", reply)
	} else {
		log.Printf(" Postprocesa la respuesta. Hubo un error : %s", err.Error())
	}

	return err
}

func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

	log.Println("======= [Interceptador de Streams en el cliente	] ", method)
	s, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return nil, err
	}
	return newWrappedStream(s), nil
}

type wrappedStream struct {
	grpc.ClientStream
}

func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del cliente] Recive un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del cliente] Envia un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ClientStream.SendMsg(m)
}

func newWrappedStream(s grpc.ClientStream) grpc.ClientStream {
	return &wrappedStream{s}
}
