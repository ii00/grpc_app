package service

import (
	"context"
	"errors"
	"grpc_app/pb"
	"log"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer is the server that provides laptop service.
type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	Store LaptopStore
}

// NewLaptopServer returns a new LaptopServer.
func NewLaptopServer(store LaptopStore) *LaptopServer {
	return &LaptopServer{Store: store}
}

func (server *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		// Check if it's a valid UUID.
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	// Some fake heavy processing.
	// time.Sleep(6 * time.Second)

	if ctx.Err() == context.Canceled {
		log.Print("request is canced")
		return nil, status.Error(codes.Canceled, "request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	// Save the laptop to storage(for now) or db.
	err := server.Store.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExist) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}
	log.Printf("saved laptop with id: %s", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

// SearchLaptop is a server-streaming RPC to search for laptops
func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("receive a search-laptop request with a filter: %v", filter)

	err := server.Store.Search(
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{
				Laptop: laptop,
			}
			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("send laptop with id: %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}