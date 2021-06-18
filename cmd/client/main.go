package main

import (
	"context"
	"flag"
	"log"

	"github.com/xiusl/pcbook/pb"
	"github.com/xiusl/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	addr := flag.String("addr", "", "the server address")
	flag.Parse()
	log.Printf("dial server: %s", *addr)

	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("cannot dial server: %v", err)
	}

	laptopClient := pb.NewLaptopServicesClient(conn)

	laptop := sample.NewLaptop()
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
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
