package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

// Interceptador unitario
func OrderUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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

//Interceptador de streams
func OrderServerStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Pre-processing
	log.Println("====== [Interceptador de Streams en el servidor] ", info.FullMethod)

	//Procesa el stream con el wrapper. El wrapper se encarga de recivir y enviar los mensajes. Este método no retorna ningun mensaje
	err := handler(srv, newWrappedStream(ss))
	if err != nil {
		log.Printf("RPC failed with error %v", err)
	}
	return err
}

//Define el wrapper que procesara los mensajes del stream
type wrappedStream struct {
	grpc.ServerStream
}

//Método que procesa los mensajes recibidos por el stream
func (w *wrappedStream) RecvMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del servidor] Se recivio un mensaje (Type: %T) at %s", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.RecvMsg(m)
}

//Método que procesa los mensajes enviados por el stream
func (w *wrappedStream) SendMsg(m interface{}) error {
	log.Printf("====== [Wrapper usado en el Stream Interceptor del servidor] Se envia un mensaje (Type: %T) at %v", m, time.Now().Format(time.RFC3339))
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}
