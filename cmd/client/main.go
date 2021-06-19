package main

import (
    "bufio"
    "context"
    "flag"
    "io"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/xiusl/pcbook/pb"
    "github.com/xiusl/pcbook/sample"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServicesClient, laptop *pb.Laptop) {
    req := &pb.CreateLaptopRequest{
        Laptop: laptop,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    res, err := laptopClient.CreateLaptop(ctx, req)
    if err != nil {
        st, ok := status.FromError(err)
        if ok && st.Code() == codes.AlreadyExists {
            log.Println("laptop already exists.")
        } else {
            log.Printf("laptop create error: %v", err)
        }
        return
    }

    log.Printf("created laptop success, id: %v", res.Id)
}

func searchTaplop(laptopClient pb.LaptopServicesClient, filter *pb.Filter) {
    log.Printf("search filter: %v", filter)

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    req := &pb.SearchLaptopRequest{
        Filter: filter,
    }
    stream, err := laptopClient.SearchLaptop(ctx, req)
    if err != nil {
        log.Fatalf("cannot search laptop: %v", err)
    }

    for {
        res, err := stream.Recv()
        if err == io.EOF {
            return
        }
        if err != nil {
            log.Fatalf("cannot receive response: %v", err)
        }

        laptop := res.GetLaptop()
        log.Print("- found: ", laptop.GetId())
        log.Print("  + brand: ", laptop.GetBrand())
        log.Print("  + name: ", laptop.GetName())
        log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
        log.Print("  + cpu min ghz: ", laptop.GetCpu().GetMinGhz())
        log.Print("  + ram: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
        log.Print("  + price: ", laptop.GetPriceUsd())

    }
}

func uploadImage(laptopClient pb.LaptopServicesClient, laptopID string, imagePath string) {
    file, err := os.Open(imagePath)
    if err != nil {
        log.Fatal("cannot open image file:", err)
    }
    defer file.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    stream, err := laptopClient.UploadImage(ctx)
    if err != nil {
        log.Fatal("cannot upload image:", err)
    }

    req := &pb.UploadImageRequest{
        Data: &pb.UploadImageRequest_Info{
            Info: &pb.ImageInfo{
                LaptopId:  laptopID,
                ImageType: filepath.Ext(imagePath),
            },
        },
    }
    err = stream.Send(req)
    if err != nil {
        log.Fatal("cannot send image info:", err)
    }

    reader := bufio.NewReader(file)
    buffer := make([]byte, 1024)

    for {
        n, err := reader.Read(buffer)
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal("cannot read chunk to buffer:", err)
        }

        req := &pb.UploadImageRequest{
            Data: &pb.UploadImageRequest_ChunkData{
                ChunkData: buffer[:n],
            },
        }

        err = stream.Send(req)
        if err != nil {
            log.Fatal("cannot send chunk data to server:", err)
        }
    }

    res, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatal("cannot receive response:", err)
    }

    log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

func testCreateLaptop(laptopClient pb.LaptopServicesClient) {
    createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServicesClient) {
    for i := 0; i < 10; i++ {
        createLaptop(laptopClient, sample.NewLaptop())
    }

    filter := &pb.Filter{
        MaxPriceUsd: 1000,
        MinCpuCores: 4,
        MinCpuGhz:   2.0,
        MinRam:      &pb.Memory{Value: 6, Unit: pb.Memory_GIGABYTE},
    }

    searchTaplop(laptopClient, filter)
}

func testUploadImage(laptopClient pb.LaptopServicesClient) {
    laptop := sample.NewLaptop()
    createLaptop(laptopClient, laptop)
    uploadImage(laptopClient, laptop.GetId(), "tmp/pc.png")
}

func main() {
    addr := flag.String("addr", "", "the server address")
    flag.Parse()
    log.Printf("dial server: %s", *addr)

    conn, err := grpc.Dial(*addr, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("cannot dial server: %v", err)
    }

    laptopClient := pb.NewLaptopServicesClient(conn)

    testUploadImage(laptopClient)
}
