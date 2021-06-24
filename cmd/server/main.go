package main

import (
    "context"
    "crypto/tls"
    "crypto/x509"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/service"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
    "google.golang.org/grpc/reflection"
)

const (
    secretKey     = "njkandsiaud"
    tokenDuration = 10 * time.Minute
)

const (
    serverCertFile   = "cert/server-cert.pem"
    serverKeyFile    = "cert/server-key.pem"
    clientCACertFile = "cert/ca-cert.pem"
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

func loadTLSCredentials() (credentials.TransportCredentials, error) {
    pemClientCA, err := ioutil.ReadFile(clientCACertFile)
    if err != nil {
        return nil, err
    }
    certPool := x509.NewCertPool()
    if !certPool.AppendCertsFromPEM(pemClientCA) {
        return nil, fmt.Errorf("cannot load the client ca file")
    }

    serverCert, err := tls.LoadX509KeyPair(serverCertFile, serverKeyFile)
    if err != nil {
        return nil, err
    }

    config := &tls.Config{
        Certificates: []tls.Certificate{serverCert},
        ClientAuth:   tls.RequireAndVerifyClientCert,
        ClientCAs:    certPool,
    }
    return credentials.NewTLS(config), nil
}

func runGRPCServer(
    authServer pb.AuthServiceServer,
    laptopServer pb.LaptopServicesServer,
    jwtManager *service.JWTManager,
    enableTLS bool,
    listener net.Listener,
) error {
    interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
    serverOptioon := []grpc.ServerOption{
        grpc.UnaryInterceptor(interceptor.Unary()),
        grpc.StreamInterceptor(interceptor.Stream()),
    }

    if enableTLS {
        tlsCredentials, err := loadTLSCredentials()
        if err != nil {
            return fmt.Errorf("cannot load  TLS credentials: %v", err)
        }
        serverOptioon = append(serverOptioon, grpc.Creds(tlsCredentials))
    }

    grpcServer := grpc.NewServer(serverOptioon...)
    pb.RegisterAuthServiceServer(grpcServer, authServer)
    pb.RegisterLaptopServicesServer(grpcServer, laptopServer)
    reflection.Register(grpcServer)

    log.Printf("Start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
    return grpcServer.Serve(listener)
}

func runRESTServer(
    authServer pb.AuthServiceServer,
    laptopServer pb.LaptopServicesServer,
    jwtManager *service.JWTManager,
    enableTLS bool,
    listener net.Listener,
) error {
    mux := runtime.NewServeMux()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    err := pb.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
    if err != nil {
        return err
    }

    err = pb.RegisterLaptopServicesHandlerServer(ctx, mux, laptopServer)
    if err != nil {
        return err
    }

    log.Printf("Start REST server at %s, TLS = %t", listener.Addr().String(), enableTLS)

    if enableTLS {
        return http.ServeTLS(listener, mux, serverCertFile, serverKeyFile)
    }

    return http.Serve(listener, mux)
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

    address := fmt.Sprintf("0.0.0.0:%s", *port)
    listener, err := net.Listen("tcp", address)
    if err != nil {
        log.Fatalf("cannot start server: %v", err)
    }

    // err = runGRPCServer(authServer, laptopServer, jwtManager, true, listener)
    // if err != nil {
    //     log.Fatalf("cannot run RPC server: %v", err)
    // }

    err = runRESTServer(authServer, laptopServer, jwtManager, false, listener)
    if err != nil {
        log.Fatalf("cannot run REST server: %v", err)
    }
}
