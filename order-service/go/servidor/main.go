package main

import (
	"context"
	"fmt"

	"io"
	"log"
	"net"
	pb "ordermgt/servidor/ecommerce"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wrapper "github.com/golang/protobuf/ptypes/wrappers"
)

const (
	port           = ":50051"
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

type server struct {
	orderMap map[string]*pb.Order
}

// Simple RPC
func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrapper.StringValue, error) {
	log.Printf("Order Added. ID : %v", orderReq.Id)
	orderMap[orderReq.Id] = *orderReq
	return &wrapper.StringValue{Value: "Order Added: " + orderReq.Id}, nil
}

// Simple RPC
func (s *server) GetOrder(ctx context.Context, orderId *wrapper.StringValue) (*pb.Order, error) {
	ord, exists := orderMap[orderId.Value]
	if exists {
		return &ord, status.New(codes.OK, "").Err()
	}

	return nil, status.Errorf(codes.NotFound, "Order does not exist. : ", orderId)

}

// Server-side Streaming RPC
//Recibimos un mensaje del cliente, y contestamos con n-mensajes enviados por el stream. Indicamos que ya no hay más mensajes en el stream cuando enviamos el mensaje nil
func (s *server) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {

	//Recibimos un mensaje del cliente, y contestamos con un stream de mensajes
	//Stream de ordenes...
	for key, order := range orderMap {
		log.Print(key, order)
		//stream de items
		for _, itemStr := range order.Items {
			log.Print(itemStr)
			//Verifica si el item es de los que buscamos
			if strings.Contains(itemStr, searchQuery.Value) {
				// Envia un mensaje por el stream al cliente
				err := stream.Send(&order)
				if err != nil {
					return fmt.Errorf("error sending message to stream : %v", err)
				}
				log.Print("Matching Order Found : " + key)
				break
			}
		}
	}
	//Termina el stream. Le estamos indicando al cliente que ya no hay más mensajes
	return nil
}

// Client-side Streaming RPC
//Recibimos n-mensajes en el stream, y cuando los hemos recibido todos contestamos con el mensaje de respuesta
func (s *server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {
	ordersStr := "Updated Order IDs : "
	for {
		//Lee un mensaje del stream
		order, err := stream.Recv()
		//Cuando el cliente cierra el stream, enviamos la respuesta desde el servidor
		if err == io.EOF {
			//Se envia el mensaje de respuesta del servidor
			return stream.SendAndClose(&wrapper.StringValue{Value: "Orders processed " + ordersStr})
		}

		if err != nil {
			return err
		}
		//Procesa el mensaje recibido en el stream
		orderMap[order.Id] = *order

		log.Printf("Order ID : %s - %s", order.Id, "Updated")
		ordersStr += order.Id + ", "
	}
}

// Bi-directional Streaming RPC
//Recibimos y enviamos n-mensajes por el stream del cliente y del servidor. Los mensajes estan entrelazados, el servidor empieza a enviar mensajes tan pronto el cliente se conecta, y el cliente puede enviar mensajes mientras el servidor esta contestando con mensajes
func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {
	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)

	//Procesa de forma indefinida
	for {
		//Recibimos un mensaje por el stream abierto desde el cliente
		orderId, err := stream.Recv()
		log.Printf("Reading Proc order : %s", orderId)
		//Si el mensaje es el EOF, el cliente significa que el cliente ha terminado de enviar sus mensajes por el stream
		if err == io.EOF {
			// Client has sent all the messages
			// Send remaining shipments
			log.Printf("EOF : %s", orderId)
			for _, shipment := range combinedShipmentMap {
				//Enviamos un mensaje desde el servidor por el stream al cliente
				if err := stream.Send(&shipment); err != nil {
					return err
				}
			}
			//Indica que el stream termina, que el servidor ya no enviara más mensajes por el stream
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}

		destination := orderMap[orderId.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]

		if found {
			ord := orderMap[orderId.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (orderMap[orderId.GetValue()].Destination), Status: "Processed!"}
			ord := orderMap[orderId.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, &ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Printf("Shipping : %v -> %v", comb.Id, len(comb.OrdersList))
				if err := stream.Send(&comb); err != nil {
					return err
				}
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}

func main() {
	initSampleData()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrderManagementServer(s, &server{})
	// Register reflection service on gRPC server.
	// reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initSampleData() {
	orderMap["102"] = pb.Order{Id: "102", Items: []string{"Google Pixel 3A", "Mac Book Pro"}, Destination: "Mountain View, CA", Price: 1800.00}
	orderMap["103"] = pb.Order{Id: "103", Items: []string{"Apple Watch S4"}, Destination: "San Jose, CA", Price: 400.00}
	orderMap["104"] = pb.Order{Id: "104", Items: []string{"Google Home Mini", "Google Nest Hub"}, Destination: "Mountain View, CA", Price: 400.00}
	orderMap["105"] = pb.Order{Id: "105", Items: []string{"Amazon Echo"}, Destination: "San Jose, CA", Price: 30.00}
	orderMap["106"] = pb.Order{Id: "106", Items: []string{"Amazon Echo", "Apple iPhone XS"}, Destination: "Mountain View, CA", Price: 300.00}
}
