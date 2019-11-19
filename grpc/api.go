package grpc

import (
	"context"
	"github.com/coschain/cos-sdk-go/utils"
	"github.com/coschain/contentos-go/rpc/pb"
	"google.golang.org/grpc"
)
const tmpChainName = "main"

type RpcWrapper struct {
	c grpcpb.ApiServiceClient

}

func NewRpcWrapper(conn *grpc.ClientConn) *RpcWrapper {
	return &RpcWrapper{c:grpcpb.NewApiServiceClient(conn)}
}

func (r *RpcWrapper) BroadcastTrx(privateKey string, op ...interface{}) (*grpcpb.BroadcastTrxResponse,error) {

	signTx, err := utils.GenerateSignedTxAndValidate(r.c, privateKey, tmpChainName,op...)

	req := &grpcpb.BroadcastTrxRequest{Transaction: signTx}
	res, err := r.c.BroadcastTrx(context.Background(),req)
	return res,err
}
