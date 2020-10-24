package nameservice

import (
	"google.golang.org/grpc/resolver"
)

const (
	ExampleScheme      = "example"
	ExampleServiceName = "lb.example.grpc.io"
)

var addrs = []string{"localhost:50051", "127.0.0.1:50051"}

// ExampleResolverBuilder Constructor de un servicio de resolución de nombres
type ExampleResolverBuilder struct{}

func (*ExampleResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &exampleResolver{
		target: target,
		cc:     cc,
		addrsStore: map[string][]string{
			ExampleServiceName: addrs, // "lb.example.grpc.io": "localhost:50051", "localhost:50052"
		},
	}
	r.start()
	return r, nil
}
func (*ExampleResolverBuilder) Scheme() string { return ExampleScheme } // "example"

type exampleResolver struct {
	target     resolver.Target
	cc         resolver.ClientConn
	addrsStore map[string][]string
}

func (r *exampleResolver) start() {
	addrStrs := r.addrsStore[r.target.Endpoint]
	addrs := make([]resolver.Address, len(addrStrs))
	for i, s := range addrStrs {
		addrs[i] = resolver.Address{Addr: s}
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}
func (*exampleResolver) ResolveNow(o resolver.ResolveNowOptions) {}
func (*exampleResolver) Close()                                  {}
