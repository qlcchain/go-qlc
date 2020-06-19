package main

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/qlcchain/go-qlc/rpc/grpc/proto"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:19746", grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	//创建客户端存根对象
	c := pb.NewLedgerAPIClient(conn)
	r, err := c.NewBlock(context.Background(), &empty.Empty{})
	if err != nil {
		fmt.Println(err)
	}
	for {
		res, err := r.Recv()
		if err != nil && err.Error() == "EOF" {
			break
		}
		if err != nil {
			log.Fatalf("%v", err)
			break
		}
		log.Printf("result:%v", res.Hash)
	}
}
