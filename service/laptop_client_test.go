package service_test

import (
	"context"
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
	laptopServer, serverAddr := startTestLaptopServer(t)
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

	other, err := laptopServer.Store.FindByID(expectedID)
	require.NoError(t, err)
	require.NotNil(t, other.Id)
	require.Equal(t, other.Id, expectedID)

	requireSameLaptop(t, other, laptop)
}

func startTestLaptopServer(t *testing.T) (*service.LaptopServer, string) {
	laptopServer := service.NewLaptopServer(service.NewInMemoryLaptopStore())

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServicesServer(grpcServer, laptopServer)

	listen, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listen)

	return laptopServer, listen.Addr().String()
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
