package main

import (
	"context"
	"flag"
	"grpc_app/pb"
	"grpc_app/sample"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	laptop.Id = ""
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// sat timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Print("laptop already exist")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return
	}

	log.Printf("created laptop with id: %s", res.Id)
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("search filter: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal("cannot receive response: ", err)
		}

		laptop := res.GetLaptop()
		log.Print("- found: ", laptop.GetId())
		log.Print("  + brand: ", laptop.GetBrand())
		log.Print("  + name: ", laptop.GetName())
		log.Print("  + cpu cores: ", laptop.GetCpu().GetNumberCores())
		log.Print("  + ram: ", laptop.GetRam().GetValue(), laptop.GetRam().GetUnit())
		log.Print("  + price: ", laptop.GetPriceUsd(), "used")
	}

}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	conn, err := grpc.Dial(*serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	laptopClient := pb.NewLaptopServiceClient(conn)
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient)
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 8,
		MinCpuGhz:   3.5,
		MinRam:      &pb.Memory{Value: 32, Unit: pb.Memory_GIGABYTE},
	}

	searchLaptop(laptopClient, filter)
}
