package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:19746", grpc.WithInsecure())
	if err != nil {
		log.Fatal("dial: ", err)
	}
	c := pb.NewTestAPIClient(conn)
	r, err := c.Version(context.Background(), &empty.Empty{})
	if err != nil {
		log.Fatal("call ", err)
	}
	fmt.Println("result, ", r)

	r2, err := c.BlockStream(context.Background(), &pb.BlockRequest{})
	if err != nil {
		log.Fatal(err)
	}
	for {
		res, err := r2.Recv()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("result ", res)
	}

}
