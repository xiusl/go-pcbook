package service_test

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xiusl/pcbook/pb"
	"github.com/xiusl/pcbook/sample"
	"github.com/xiusl/pcbook/serializer"
	"github.com/xiusl/pcbook/service"
	"google.golang.org/grpc"
)

func TestClientCreateLaptop(t *testing.T) {

	laptopStore := service.NewInMemoryLaptopStore()
	serverAddr := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddr)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id

	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	resp, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, resp.Id, expectedID)

	other, err := laptopStore.FindByID(expectedID)
	require.NoError(t, err)
	require.NotNil(t, other.Id)
	require.Equal(t, other.Id, expectedID)

	requireSameLaptop(t, other, laptop)
}

func TestClientSearchLaptop(t *testing.T) {

	filter := &pb.Filter{
		MaxPriceUsd: 1000,
		MinCpuCores: 4,
		MinCpuGhz:   2.0,
		MinRam:      &pb.Memory{Value: 6, Unit: pb.Memory_GIGABYTE},
	}

	store := service.NewInMemoryLaptopStore()
	expectedIDs := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 1500
		case 1:
			laptop.Cpu.NumberCores = 1
		case 2:
			laptop.Cpu.MinGhz = 1.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 1, Unit: pb.Memory_GIGABYTE}
		case 4:
			laptop.PriceUsd = 900
			laptop.Cpu.NumberCores = 8
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 888
			laptop.Cpu.NumberCores = 16
			laptop.Cpu.MinGhz = 3.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIDs[laptop.Id] = true
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

	serverAddr := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, serverAddr)

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIDs, res.GetLaptop().GetId())
		found += 1
	}
	require.Equal(t, found, len(expectedIDs))
}

func startTestLaptopServer(t *testing.T, laptopstroe service.LaptopStore, imageStore service.ImageStore) string {
	laptopServer := service.NewLaptopServer(laptopstroe, imageStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServicesServer(grpcServer, laptopServer)

	listen, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listen)

	return listen.Addr().String()
}

func newTestLaptopClient(t *testing.T, addr string) pb.LaptopServicesClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	require.NoError(t, err)
	return pb.NewLaptopServicesClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1, laptop2 *pb.Laptop) {
	json1, err := serializer.ConvertProtobufToJSON(laptop1)
	require.NoError(t, err)

	json2, err := serializer.ConvertProtobufToJSON(laptop2)
	require.NoError(t, err)

	require.Equal(t, json1, json2)
}
