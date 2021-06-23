package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net"

    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/service"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

func unaryInterceptor(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {

    log.Println("---> unary interceptor: ", info.FullMethod)
    return handler(ctx, req)
}

func streamInterceptor(
    srv interface{},
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    log.Println("---> stream interceptor: ", info.FullMethod)
    return handler(srv, ss)
}

func main() {
    port := flag.String("port", "", "server port")
    flag.Parse()
    log.Printf("start server on port: %s", *port)

    laptopStore := service.NewInMemoryLaptopStore()
    imageStore := service.NewDiskImageStore("img")
    ratingStore := service.NewInMemoryRatingStore()
    laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(unaryInterceptor),
        grpc.StreamInterceptor(streamInterceptor),
    )
    pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
    reflection.Register(grpcServer)

    address := fmt.Sprintf("0.0.0.0:%s", *port)
    listener, err := net.Listen("tcp", address)
    if err != nil {
        log.Fatalf("cannot start server: %v", err)
    }

    err = grpcServer.Serve(listener)
    if err != nil {
        log.Fatalf("cannot start server: %v", err)
    }

}
