package logica

import (
	"context"
	"fmt"
	pb "interceptors/servidor/ecommerce"
	"io"
	"log"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	orderBatchSize = 3
)

//Server implementa la lógica de negocio del servicio RPC
type Server struct {
	orderMap map[string]*pb.Order
}

//Construye construye el servicio
func Construye(ordenes map[string]*pb.Order) *Server {
	return &Server{orderMap: ordenes}
}

//AddOrder añade una orden (Simple RPC)
func (s *Server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrappers.StringValue, error) {

	//Obtiene los metadatos del contexto
	md, metadataAvailable := metadata.FromIncomingContext(ctx)

	//Si no hay metadatos devuelve un error
	if !metadataAvailable {
		return nil, status.Errorf(codes.DataLoss, "UnaryEcho: failed to get metadata")
	}

	//Si hay metadatos, recupera el metadato timestamp, y los detalles asociados
	if t, ok := md["timestamp"]; ok {
		fmt.Printf("timestamp from metadata:\n")
		for i, e := range t {
			fmt.Printf("====> Metadata %d. %s\n", i, e)
		}
	}

	//Creamos metadatos
	header := metadata.New(map[string]string{"location": "San Jose", "timestamp": time.Now().Format(time.StampNano)})
	//Los añadimos a la cabecera de respuesta
	grpc.SendHeader(ctx, header)

	if orderReq.Id == "-1" {
		log.Printf("Order ID is invalid! -> Received Order ID %s", orderReq.Id)

		//Crea un error
		errorStatus := status.New(codes.InvalidArgument, "Invalid information received")

		//Podemos añadir detalles al error
		ds, err := errorStatus.WithDetails(
			&epb.BadRequest_FieldViolation{
				Field:       "ID",
				Description: fmt.Sprintf("Order ID received is not valid %s : %s", orderReq.Id, orderReq.Description),
			},
		)
		if err != nil {
			return nil, errorStatus.Err()
		}

		return nil, ds.Err()
	} else {
		s.orderMap[orderReq.Id] = orderReq
		log.Println("Order : ", orderReq.Id, " -> Added")
		return &wrappers.StringValue{Value: "Order Added: " + orderReq.Id}, nil
	}
}

//GetOrder busca una orden (Simple RPC)
func (s *Server) GetOrder(ctx context.Context, orderId *wrappers.StringValue) (*pb.Order, error) {
	ord := s.orderMap[orderId.Value]
	return ord, nil
}

//SearchOrders busca ordenes (Server-side Streaming RPC)
func (s *Server) SearchOrders(searchQuery *wrappers.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {

	//Añade los metadatos al trailer
	defer func() {
		trailer := metadata.Pairs("timestamp", time.Now().Format(time.StampNano))
		stream.SetTrailer(trailer)
	}()

	//Añade los metadatos al header
	header := metadata.New(map[string]string{"location": "MTV", "timestamp": time.Now().Format(time.StampNano)})
	stream.SendHeader(header)

	for key, order := range s.orderMap {
		for _, itemStr := range order.Items {
			if strings.Contains(itemStr, searchQuery.Value) {
				// Send the matching orders in a stream
				log.Print("Matching Order Found : "+key, " -> Writing Order to the stream ... ")
				stream.Send(order)
				break
			}
		}
	}
	return nil
}

//UpdateOrders actualiza ordenes (Client-side Streaming RPC)
func (s *Server) UpdateOrders(stream pb.OrderManagement_UpdateOrdersServer) error {

	ordersStr := "Updated Order IDs : "
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			// Finished reading the order stream.
			return stream.SendAndClose(&wrappers.StringValue{Value: "Orders processed " + ordersStr})
		}
		// Update order
		s.orderMap[order.Id] = order

		log.Printf("Order ID ", order.Id, ": Updated")
		ordersStr += order.Id + ", "
	}
}

//ProcessOrders procesa ordenes (Bi-directional Streaming RPC)
func (s *Server) ProcessOrders(stream pb.OrderManagement_ProcessOrdersServer) error {

	batchMarker := 1
	var combinedShipmentMap = make(map[string]pb.CombinedShipment)
	for {
		orderID, err := stream.Recv()
		log.Println("Reading Proc order ... ", orderID)
		if err == io.EOF {
			// Client has sent all the messages
			// Send remaining shipments

			log.Println("EOF ", orderID)

			for _, comb := range combinedShipmentMap {
				stream.Send(&comb)
			}
			return nil
		}
		if err != nil {
			log.Println(err)
			return err
		}

		destination := s.orderMap[orderID.GetValue()].Destination
		shipment, found := combinedShipmentMap[destination]

		if found {
			ord := s.orderMap[orderID.GetValue()]
			shipment.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[destination] = shipment
		} else {
			comShip := pb.CombinedShipment{Id: "cmb - " + (s.orderMap[orderID.GetValue()].Destination), Status: "Processed!"}
			ord := s.orderMap[orderID.GetValue()]
			comShip.OrdersList = append(shipment.OrdersList, ord)
			combinedShipmentMap[destination] = comShip
			log.Print(len(comShip.OrdersList), comShip.GetId())
		}

		if batchMarker == orderBatchSize {
			for _, comb := range combinedShipmentMap {
				log.Print("Shipping : ", comb.Id, " -> ", len(comb.OrdersList))
				stream.Send(&comb)
			}
			batchMarker = 0
			combinedShipmentMap = make(map[string]pb.CombinedShipment)
		} else {
			batchMarker++
		}
	}
}
