package utils

import (
	"context"
	"errors"
	"github.com/coschain/contentos-go/common"
	"github.com/coschain/contentos-go/prototype"
	"github.com/coschain/contentos-go/rpc/pb"
	"hash/crc32"
	"math/rand"
	"time"
)

type ChainId string

const (
	Main ChainId = "main"
	Test ChainId = "test"
	Dev ChainId = "dev"
)

var GlobalChainId ChainId

func GenerateSignedTxAndValidate(client grpcpb.ApiServiceClient, privateKey string, chainName string, ops ...interface{}) (*prototype.SignedTransaction, error) {
	chainId := prototype.ChainId{ Value:common.GetChainIdByName(chainName)}
	return GenerateSignedTxAndValidate2(client, privateKey, chainId, ops...)
}

func GenerateSignedTxAndValidate2(client grpcpb.ApiServiceClient, privateKey string, chainId prototype.ChainId, ops ...interface{}) (*prototype.SignedTransaction, error) {
	privKey, err := prototype.PrivateKeyFromWIF(privateKey)
	if err != nil {
		return nil, err
	}
	return GenerateSignedTxAndValidate3(client, privKey, chainId, ops...)
}

func GenerateSignedTxAndValidate3(client grpcpb.ApiServiceClient, privKey *prototype.PrivateKeyType, chainId prototype.ChainId, ops ...interface{}) (*prototype.SignedTransaction, error) {
	chainState, err := GetChainState(client)
	if err != nil {
		return nil, err
	}
	return GenerateSignedTxAndValidate4(chainState.Dgpo, 30, privKey, chainId, ops...)
}

func GetChainState(client grpcpb.ApiServiceClient) (*grpcpb.ChainState, error) {
	req := &grpcpb.NonParamsRequest{}
	resp, err := client.GetChainState(context.Background(), req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("response == nil, err == nil")
	}
	return resp.State, nil
}

func GenerateSignedTxAndValidate4(dgp *prototype.DynamicProperties, expiration uint32, privKey *prototype.PrivateKeyType, chainId prototype.ChainId, ops ...interface{}) (*prototype.SignedTransaction, error) {
	refBlockPrefix := common.TaposRefBlockPrefix(dgp.HeadBlockId.Hash)
	// occupant implement
	refBlockNum := common.TaposRefBlockNum(dgp.HeadBlockNumber)
	tx := &prototype.Transaction{RefBlockNum: refBlockNum, RefBlockPrefix: refBlockPrefix, Expiration: &prototype.TimePointSec{UtcSeconds: dgp.Time.UtcSeconds + expiration}}
	for _, op := range ops {
		tx.AddOperation(op)
	}

	signTx := prototype.SignedTransaction{Trx: tx}

	res := signTx.Sign(privKey, chainId)
	signTx.Signature = &prototype.SignatureType{Sig: res}

	if err := signTx.Validate(); err != nil {
		return nil, err
	}

	return &signTx, nil
}

func GenerateUUID(content string) uint64 {
	crc32q := crc32.MakeTable(0xD5828281)
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	randContent := content + string(r.Intn(100000))
	return uint64(time.Now().Unix())*uint64(1e9) + uint64(crc32.Checksum([]byte(randContent), crc32q))
}
