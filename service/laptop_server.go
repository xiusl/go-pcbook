package service

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/xiusl/pcbook/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer 提供 laptop 服务的服务器
type LaptopServer struct {
	Store LaptopStore
}

// NewLaptopServer 创建一个 laptop 服务器
func NewLaptopServer(stroe LaptopStore) *LaptopServer {
	return &LaptopServer{
		Store: stroe,
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

	err := server.Store.Save(laptop)
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
