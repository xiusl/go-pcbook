package service

import (
    "bytes"
    "context"
    "errors"
    "fmt"
    "io"
    "log"

    "github.com/google/uuid"
    "github.com/xiusl/pcbook/pb"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// 1 mb
const maxImageSize = 1 << 20

// LaptopServer 提供 laptop 服务的服务器
type LaptopServer struct {
    laptopStore LaptopStore
    imageStore  ImageStore
    ratingStore RatingStore
}

// NewLaptopServer 创建一个 laptop 服务器
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
    return &LaptopServer{
        laptopStore: laptopStore,
        imageStore:  imageStore,
        ratingStore: ratingStore,
    }
}

// CreateLaptop 实现创建 laptop 的方法
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
    laptop := req.GetLaptop()
    log.Printf("receive a create-laptop request with id:%s.\n", laptop.Id)
    // 检查 UUID
    if len(laptop.Id) > 0 {
        _, err := uuid.Parse(laptop.Id)
        if err != nil {
            return nil, status.Errorf(codes.InvalidArgument, "laptap ID is not a valid UUID: %v", err)
        }
    } else {
        id, err := uuid.NewRandom()
        if err != nil {
            return nil, status.Errorf(codes.Internal, "cannot generate a new UUID: %v", err)
        }
        laptop.Id = id.String()
    }

    if ctx.Err() == context.Canceled {
        log.Print("context is canceled")
        return nil, fmt.Errorf("context is canceled")
    }

    if ctx.Err() == context.DeadlineExceeded {
        log.Print("deadline is exceeded")
        return nil, fmt.Errorf("deadline is exceeded")
    }

    err := server.laptopStore.Save(laptop)
    if err != nil {
        code := codes.Internal
        if errors.Is(err, ErrAlreadyExists) {
            code = codes.AlreadyExists
        }
        return nil, status.Errorf(code, "cannot save the laptop to store: %v", err)
    }

    res := &pb.CreateLaptopResponse{Id: laptop.Id}
    return res, nil
}

func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopServices_SearchLaptopServer) error {
    filter := req.GetFilter()
    log.Printf("receive a search-laptop request with filter: %v", filter)

    err := server.laptopStore.Search(stream.Context(), filter, func(laptop *pb.Laptop) error {
        res := &pb.SearchLaptopResponse{
            Laptop: laptop,
        }
        err := stream.Send(res)
        if err != nil {
            return err
        }
        return nil
    })

    if err != nil {
        return status.Errorf(codes.Internal, "unexpected error: %v", err)
    }
    return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopServices_UploadImageServer) error {

    // 读取请求信息
    req, err := stream.Recv()
    if err != nil {
        log.Print("cannot receive image info", err)
        return status.Error(codes.Unknown, "cannot receive image info")
    }

    laptapID := req.GetInfo().GetLaptopId()
    imageType := req.GetInfo().GetImageType()
    log.Printf("receive an upload-image request for laptop %s with image type %s", laptapID, imageType)

    // 获取需要存储图片的便携电脑
    laptap, err := server.laptopStore.FindByID(laptapID)
    if err != nil {
        log.Print("cannot find the laptop", err)
        return status.Error(codes.Internal, "cannot find the laptop")
    }
    if laptap == nil {
        log.Printf("laptop %s doesn't exist", laptapID)
        return status.Errorf(codes.InvalidArgument, "laptop %s doesn't exist", laptapID)
    }

    imageData := bytes.Buffer{}
    imageSize := 0

    // 开始从请求中不断接受数据
    for {
        log.Print("wait to receive more data")

        // 对上下文进行判断
        if stream.Context().Err() == context.Canceled {
            log.Print("context is canceled")
            return fmt.Errorf("context is canceled")
        }

        if stream.Context().Err() == context.DeadlineExceeded {
            log.Print("deadline is exceeded")
            return fmt.Errorf("deadline is exceeded")
        }

        // 接受请求
        req, err := stream.Recv()
        if err == io.EOF {
            log.Print("no more data")
            break
        }
        if err != nil {
            log.Printf("cannot receive chunk data: %v", err)
            return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
        }

        // 获取请求中的图像数据，分块的
        chunk := req.GetChunkData()
        size := len(chunk)

        imageSize += size
        if imageSize > maxImageSize {
            log.Printf("image to large %v > %v", imageSize, maxImageSize)
            return status.Errorf(codes.InvalidArgument, "image to large %v > %v", imageSize, maxImageSize)
        }

        // 将分块的数据写入到 imageData 中
        _, err = imageData.Write(chunk)
        if err != nil {
            log.Printf("cannot write chunk data: %v", err)
            return status.Errorf(codes.Internal, "cannot write chunk data: %v", err)
        }
    }

    // 调用存储，保存图片
    imageID, err := server.imageStore.Save(laptapID, imageType, imageData)
    if err != nil {
        log.Printf("cannot save image to file: %v", err)
        return status.Errorf(codes.Internal, "cannot save image to file: %v", err)
    }

    res := &pb.UploadImageResponse{
        Id:   imageID,
        Size: uint32(imageSize),
    }

    // 发送结束响应并关闭流
    err = stream.SendAndClose(res)
    if err != nil {
        log.Printf("cannot send the response: %v", err)
        return status.Errorf(codes.Internal, "cannot send the response: %v", err)
    }

    log.Printf("saved image with id: %s, size: %d", imageID, imageSize)
    return nil
}

// RateLaptop 对 laptop 进行打分
func (server *LaptopServer) RateLaptop(stream pb.LaptopServices_RateLaptopServer) error {
    for {
        // 对上下文进行判断
        if stream.Context().Err() == context.Canceled {
            log.Print("context is canceled")
            return fmt.Errorf("context is canceled")
        }

        if stream.Context().Err() == context.DeadlineExceeded {
            log.Print("deadline is exceeded")
            return fmt.Errorf("deadline is exceeded")
        }

        req, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Printf("cannot receive stream data: %v", err)
            return status.Errorf(codes.Unknown, "cannot receive stream data: %v", err)
        }

        laptopID := req.GetLaptopId()
        scroe := req.GetScore()

        laptap, err := server.laptopStore.FindByID(laptopID)
        if err != nil {
            log.Print("cannot find the laptop", err)
            return status.Error(codes.Internal, "cannot find the laptop")
        }
        if laptap == nil {
            log.Printf("laptop %s doesn't exist", laptopID)
            return status.Errorf(codes.InvalidArgument, "laptop %s doesn't exist", laptopID)
        }

        rating, err := server.ratingStore.Add(laptopID, scroe)
        if err != nil {
            log.Printf("cannot add the score to store %v.", err)
            return status.Errorf(codes.Internal, "cannot add the score to store %v.", err)
        }

        res := &pb.RateLaptopResponse{
            LaptopId:     laptopID,
            RatedCount:   rating.Count,
            AverageScote: rating.Sum / float64(rating.Count),
        }

        err = stream.Send(res)
        if err != nil {
            log.Printf("cannot send the response: %v", err)
            return status.Errorf(codes.Internal, "cannot send the response: %v", err)
        }
    }
    return nil
}
