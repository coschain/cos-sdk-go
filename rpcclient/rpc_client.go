package rpcclient

import (
	"github.com/coschain/contentos-go/rpc/pb"
	"google.golang.org/grpc"
	"sync"
)

var once sync.Once
var rpcClient grpcpb.ApiServiceClient
var oldConn *grpc.ClientConn

func GetRpc() grpcpb.ApiServiceClient {
	once.Do(func() {
		if err := ConnectRpc("localhost:8888"); err != nil {
			panic(err)
		}
	})
	return rpcClient
}

func ConnectRpc(ip string) error {
	if oldConn != nil {
		oldConn.Close()
	}

	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		return err
	}
	rpcClient = grpcpb.NewApiServiceClient(conn)
	oldConn = conn

	return nil
}