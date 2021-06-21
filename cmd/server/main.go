package main

import (
    "flag"
    "fmt"
    "log"
    "net"

    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/service"
    "google.golang.org/grpc"
)

func main() {
    port := flag.String("port", "", "server port")
    flag.Parse()
    log.Printf("start server on port: %s", *port)

    laptopStore := service.NewInMemoryLaptopStore()
    imageStore := service.NewDiskImageStore("img")
    ratingStore := service.NewInMemoryRatingStore()
    laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
    grpcServer := grpc.NewServer()
    pb.RegisterLaptopServicesServer(grpcServer, laptopServer)

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
