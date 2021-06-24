package main

import (
    "flag"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/xiusl/pcbook/client"
    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/sample"
    "google.golang.org/grpc"
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
    username        = "user1"
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

func main() {
    addr := flag.String("addr", "", "the server address")
    flag.Parse()
    log.Printf("dial server: %s", *addr)

    conn, err := grpc.Dial(*addr, grpc.WithInsecure())
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
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(interceptor.Unary()),
        grpc.WithStreamInterceptor(interceptor.Stream()),
    )
    if err != nil {
        log.Fatalf("cannot dial server2: %v", err)
    }

    laptopClient := client.NewLaptopClient(conn1)

    testRatingLaptop(laptopClient)
}
