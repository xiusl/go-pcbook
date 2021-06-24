package main

import (
    "crypto/tls"
    "crypto/x509"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "strings"
    "time"

    "github.com/xiusl/pcbook/client"
    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/sample"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

func testCreateLaptop(laptopClient *client.LaptopClient) {
    laptopClient.CreateLaptop(sample.NewLaptop())
}

func testSearchLaptop(laptopClient *client.LaptopClient) {
    for i := 0; i < 10; i++ {
        laptopClient.CreateLaptop(sample.NewLaptop())
    }

    filter := &pb.Filter{
        MaxPriceUsd: 1000,
        MinCpuCores: 4,
        MinCpuGhz:   2.0,
        MinRam:      &pb.Memory{Value: 6, Unit: pb.Memory_GIGABYTE},
    }

    laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient *client.LaptopClient) {
    laptop := sample.NewLaptop()
    laptopClient.CreateLaptop(laptop)
    laptopClient.UploadImage(laptop.GetId(), "tmp/pc.png")
}

func testRatingLaptop(laptopClient *client.LaptopClient) {
    n := 3
    laptopIDs := make([]string, n)

    for i := 0; i < n; i++ {
        laptop := sample.NewLaptop()
        laptopClient.CreateLaptop(laptop)
        laptopIDs[i] = laptop.GetId()
    }

    scores := make([]float64, n)
    for {
        fmt.Println("rate laptop (y/n)?:")
        var ans string
        fmt.Scan(&ans)

        if strings.ToLower(ans) != "y" {
            break
        }

        for i := 0; i < n; i++ {
            scores[i] = sample.RandomLaptopScore()
        }

        err := laptopClient.RateLaptop(laptopIDs, scores)
        if err != nil {
            log.Fatal(err)
        }
    }

}

const (
    username        = "admin"
    password        = "abc"
    refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
    const latopServicePath = "/xiusl.pcbook.LaptopServices/"
    return map[string]bool{
        latopServicePath + "CreateLaptop": true,
        latopServicePath + "UploadImage":  true,
        latopServicePath + "RateLaptop":   true,
    }
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
    pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
    if err != nil {
        return nil, err
    }
    certPool := x509.NewCertPool()
    if !certPool.AppendCertsFromPEM(pemServerCA) {
        return nil, fmt.Errorf("cannot load the server ca file")
    }

    clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
    if err != nil {
        return nil, err
    }

    config := &tls.Config{
        Certificates: []tls.Certificate{clientCert},
        RootCAs:      certPool,
    }
    return credentials.NewTLS(config), nil
}

func main() {
    addr := flag.String("addr", "", "the server address")
    enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
    flag.Parse()
    log.Printf("dial server: %s", *addr)

    transportOption := grpc.WithInsecure()
    if *enableTLS {
        tlsCredentials, err := loadTLSCredentials()
        if err != nil {
            log.Fatalf("cannot load  TLS credentials: %v", err)
        }
        transportOption = grpc.WithTransportCredentials(tlsCredentials)
    }

    conn, err := grpc.Dial(*addr, transportOption)
    if err != nil {
        log.Fatalf("cannot dial server: %v", err)
    }

    authClient := client.NewAuthClient(conn, username, password)
    interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
    if err != nil {
        log.Fatalf("cannot create auth interceptor: %v", err)
    }

    conn1, err := grpc.Dial(
        *addr,
        transportOption,
        grpc.WithUnaryInterceptor(interceptor.Unary()),
        grpc.WithStreamInterceptor(interceptor.Stream()),
    )
    if err != nil {
        log.Fatalf("cannot dial server2: %v", err)
    }

    laptopClient := client.NewLaptopClient(conn1)

    testRatingLaptop(laptopClient)
}
