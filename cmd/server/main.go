package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "time"

    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/service"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
)

const (
    secretKey     = "njkandsiaud"
    tokenDuration = 10 * time.Minute
)

func seedUser(userStroe service.UserStore) error {
    err := createUser(userStroe, "admin", "abc", "admin")
    if err != nil {
        return err
    }
    return createUser(userStroe, "user1", "abc", "user")
}

func createUser(userStroe service.UserStore, username, password, role string) error {
    user, err := service.NewUser(username, password, role)
    if err != nil {
        return err
    }
    return userStroe.Save(user)
}

func accessibleRoles() map[string][]string {
    const latopServicePath = "/xiusl.pcbook.LaptopServices/"
    return map[string][]string{
        latopServicePath + "CreateLaptop": {"admin"},
        latopServicePath + "UploadImage":  {"admin"},
        latopServicePath + "RateLaptop":   {"admin", "user"},
    }
}

func main() {
    port := flag.String("port", "", "server port")
    flag.Parse()
    log.Printf("start server on port: %s", *port)

    userStore := service.NewInMemoryUserStore()
    if err := seedUser(userStore); err != nil {
        log.Fatal("cannot create seed users: %w", err)
    }
    jwtManager := service.NewJWTManager(secretKey, tokenDuration)

    authServer := service.NewAuthServer(userStore, jwtManager)

    laptopStore := service.NewInMemoryLaptopStore()
    imageStore := service.NewDiskImageStore("img")
    ratingStore := service.NewInMemoryRatingStore()
    laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

    interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(interceptor.Unary()),
        grpc.StreamInterceptor(interceptor.Stream()),
    )
    pb.RegisterAuthServiceServer(grpcServer, authServer)
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
