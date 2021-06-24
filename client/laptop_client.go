package client

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/xiusl/pcbook/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// LaptopClient 调用便携电脑的 RPC 客户端
type LaptopClient struct {
    server pb.LaptopServicesClient
}

// NewLaptopClien 创建一个新的客户端
func NewLaptopClient(cc *grpc.ClientConn) *LaptopClient {
    server := pb.NewLaptopServicesClient(cc)
    return &LaptopClient{server}
}

// CreateLaptop 创建一个便携电脑
func (clien *LaptopClient) CreateLaptop(laptop *pb.Laptop) {
    req := &pb.CreateLaptopRequest{
        Laptop: laptop,
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    res, err := clien.server.CreateLaptop(ctx, req)
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

// SearchLaptop 搜索指定的便携电脑
func (client *LaptopClient) SearchLaptop(filter *pb.Filter) {
    log.Printf("search filter: %v", filter)

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    req := &pb.SearchLaptopRequest{
        Filter: filter,
    }
    stream, err := client.server.SearchLaptop(ctx, req)
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

// UploadImage 为指定的便携电脑上传图片
func (clien *LaptopClient) UploadImage(laptopID, imagePath string) {
    // 打开文件
    file, err := os.Open(imagePath)
    if err != nil {
        log.Fatal("cannot open image file:", err)
    }
    defer file.Close()

    // 创建一个带有 2s 超时的上下文
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    // 调用客户端，开启一个请求流
    stream, err := clien.server.UploadImage(ctx)
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

    // 先发送图片基本的信息
    err = stream.Send(req)
    if err != nil {
        log.Fatal("cannot send image info:", err)
    }

    reader := bufio.NewReader(file)
    // 创建一个 1024 byte 的二级制数据块
    buffer := make([]byte, 1024)

    for {
        // 每次读取 1mb 数据
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

        // 发送数据
        err = stream.Send(req)
        if err != nil {
            log.Fatal("cannot send chunk data to server:", err)
        }
    }

    // 关闭并接收响应
    res, err := stream.CloseAndRecv()
    if err != nil {
        log.Fatal("cannot receive response:", err)
    }

    log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

// RateLaptop 为便携电脑打分
func (client *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    stream, err := client.server.RateLaptop(ctx)
    if err != nil {
        log.Fatal("cannot rate laptop:", err)
    }

    withResp := make(chan error)
    go func() {
        for {
            res, err := stream.Recv()

            if err == io.EOF {
                log.Print("no more response")
                withResp <- nil
                break
            }
            if err != nil {
                withResp <- fmt.Errorf("cannot receive stream response %w", err)
                return
            }

            log.Print("receive response: ", res)
        }
    }()

    for i, laptopID := range laptopIDs {
        req := &pb.RateLaptopRequest{
            LaptopId: laptopID,
            Score:    scores[i],
        }

        err = stream.Send(req)
        if err != nil {
            return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
        }

        log.Print("send request", req)
    }

    err = stream.CloseSend()
    if err != nil {
        return fmt.Errorf("cannot close send: %v", err)
    }

    err = <-withResp
    return err
}
