package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func OrderUnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
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

func ClientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {

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
