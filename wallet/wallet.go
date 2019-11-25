package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"github.com/coschain/contentos-go/prototype"
	"github.com/coschain/contentos-go/rpc/pb"
	"github.com/coschain/cos-sdk-go/account"
	"github.com/coschain/cos-sdk-go/rpcclient"
	"github.com/coschain/cos-sdk-go/utils"
	"github.com/kataras/go-errors"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"context"
)

type EncryptKeyStore struct {
	CipherText string // encrypted privkey
	Iv         string // the iv
	Mac        string // the mac of passphrase
}

type QueryRpc interface {
	GetAccountByName(name string) error
}

type Wallet interface {
	QueryRpc

	Open(path string, password string) error
	Add(name, privateKey string) error
	Remove(name string) error
	Account(name string) *account.Account // name -> account
	Close()
}

func (w *walletImpl) QueryTableContent(owner,contract,table,field string, count uint32, reverse bool) (*grpcpb.TableContentResponse,error) {
	req := &grpcpb.GetTableContentRequest{
		Owner:owner,
		Contract:contract,
		Table:table,
		Field:field,
		Count:count,
		Reverse:reverse,
	}
	return rpcclient.GetRpc().QueryTableContent(context.Background(),req)
}

func (w *walletImpl) GetAccountByName(name string) (*grpcpb.AccountResponse,error) {
	req := &grpcpb.GetAccountByNameRequest{AccountName: &prototype.AccountName{Value: name}}
	return rpcclient.GetRpc().GetAccountByName(context.Background(), req)
}

func (w *walletImpl) GetFollowerListByName(name string, pageSize uint32) (*PageManager,error) {

	start := &prototype.FollowerCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:0},
		Follower:prototype.NewAccountName(""),
	}
	end := &prototype.FollowerCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:math.MaxUint32},
		Follower:prototype.NewAccountName(""),
	}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetFollowerListByNameRequest{
			Start:page.Start.(*prototype.FollowerCreatedOrder),
			End:page.End.(*prototype.FollowerCreatedOrder),
			Limit:page.Limit,
			LastOrder:page.LastOrder.(*prototype.FollowerCreatedOrder),
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetFollowerListByName(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.FollowerList) == 0 {
			return nil,nil,errors.New("empty result")
		}
	    lastOrder := res.FollowerList[len(res.FollowerList)-1].CreateOrder

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetFollowingListByName(name string, pageSize uint32) (*PageManager,error) {
	start := &prototype.FollowingCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:0},
		Following:prototype.NewAccountName(""),
	}
	end := &prototype.FollowingCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:math.MaxUint32},
		Following:prototype.NewAccountName(""),
	}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetFollowingListByNameRequest{
			Start:page.Start.(*prototype.FollowingCreatedOrder),
			End:page.End.(*prototype.FollowingCreatedOrder),
			Limit:page.Limit,
			LastOrder:page.LastOrder.(*prototype.FollowingCreatedOrder),
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetFollowingListByName(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.FollowingList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.FollowingList[len(res.FollowingList)-1].CreateOrder

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetFollowCountByName(name string) (*grpcpb.GetFollowCountByNameResponse,error) {
	req := &grpcpb.GetFollowCountByNameRequest{
		AccountName:prototype.NewAccountName(name),
	}
	return rpcclient.GetRpc().GetFollowCountByName(context.Background(),req)
}

func (w *walletImpl) GetBlockProducerList(size uint32) (*grpcpb.GetBlockProducerListResponse,error) {
	req := &grpcpb.GetBlockProducerListRequest{
		Start:prototype.NewAccountName(""),
		Limit:size,
	}
	return rpcclient.GetRpc().GetBlockProducerList(context.Background(),req)
}

func (w *walletImpl) GetPostListByCreated(startTime uint32, endTime uint32, limit uint32) (*grpcpb.GetPostListByCreatedResponse,error) {
	req := &grpcpb.GetPostListByCreatedRequest{
		Start:&prototype.PostCreatedOrder{Created:prototype.NewTimePointSec(startTime)},
		End:&prototype.PostCreatedOrder{Created:prototype.NewTimePointSec(endTime)},
		Limit:limit,
	}
	return rpcclient.GetRpc().GetPostListByCreated(context.Background(),req)
}

func (w *walletImpl) GetReplyListByPostId(postid uint64, startTime uint32, endTime uint32, limit uint32) (*grpcpb.GetReplyListByPostIdResponse,error) {
	req := &grpcpb.GetReplyListByPostIdRequest{
		Start:&prototype.ReplyCreatedOrder{ParentId:postid,Created:prototype.NewTimePointSec(0)},
		End:&prototype.ReplyCreatedOrder{ParentId:postid,Created:prototype.NewTimePointSec(math.MaxUint32)},
		Limit:limit,
	}
	return rpcclient.GetRpc().GetReplyListByPostId(context.Background(),req)
}

func (w *walletImpl) GetBlockTransactionsByNum(blockNum uint32) (*grpcpb.GetBlockTransactionsByNumResponse,error) {
	req := &grpcpb.GetBlockTransactionsByNumRequest{BlockNum:blockNum}
	return rpcclient.GetRpc().GetBlockTransactionsByNum(context.Background(),req)
}

func (w *walletImpl) GetChainState() (*grpcpb.GetChainStateResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetChainState(context.Background(),req)
}

func (w *walletImpl) GetBlockList(start, end uint64, limit uint32) (*grpcpb.GetBlockListResponse,error) {
	req := &grpcpb.GetBlockListRequest{
		Start:start,
		End:end,
		Limit:limit,
	}
	return rpcclient.GetRpc().GetBlockList(context.Background(),req)
}

func (w *walletImpl) GetSignedBlock(blockNum uint64) (*grpcpb.GetSignedBlockResponse,error) {
	req := &grpcpb.GetSignedBlockRequest{Start:blockNum}
	return rpcclient.GetRpc().GetSignedBlock(context.Background(),req)
}

func (w *walletImpl) GetAccountListByBalance(startCoin,endCoin uint64, pageSize uint32) (*PageManager,error) {
	 start := prototype.NewCoin(startCoin)
	 end := prototype.NewCoin(endCoin)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetAccountListByBalanceRequest{
			Start:page.Start.(*prototype.Coin),
			End:page.End.(*prototype.Coin),
			LastAccount:page.LastOrder.(*grpcpb.AccountInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetAccountListByBalance(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.List) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.List[len(res.List)-1].Info

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetDailyTotalTrxInfo(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetDailyTotalTrxRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastInfo:page.LastOrder.(*grpcpb.DailyTotalTrx),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetDailyTotalTrxInfo(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.List) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.List[len(res.List)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetTrxInfoById(trxId *prototype.Sha256) (*grpcpb.GetTrxInfoByIdResponse,error) {
	req := &grpcpb.GetTrxInfoByIdRequest{TrxId:trxId}
	return rpcclient.GetRpc().GetTrxInfoById(context.Background(),req)
}

func (w *walletImpl) GetTrxListByTime(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetTrxListByTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastInfo:page.LastOrder.(*grpcpb.TrxInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetTrxListByTime(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.List) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.List[len(res.List)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetPostListByCreateTime(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetPostListByCreateTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastPost:page.LastOrder.(*grpcpb.PostResponse),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetPostListByCreateTime(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.PostedList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.PostedList[len(res.PostedList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetPostListByName(name string,pageSize uint32) (*PageManager,error) {
	start := &prototype.UserPostCreateOrder{Author:prototype.NewAccountName(name),Create:prototype.NewTimePointSec(0)}
	end := &prototype.UserPostCreateOrder{Author:prototype.NewAccountName(name),Create:prototype.NewTimePointSec(math.MaxUint32)}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetPostListByNameRequest{
			Start:page.Start.(*prototype.UserPostCreateOrder),
			End:page.End.(*prototype.UserPostCreateOrder),
			LastPost:page.LastOrder.(*grpcpb.PostResponse),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetPostListByName(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.PostedList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.PostedList[len(res.PostedList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) TrxStatByHour(hours uint32) (*grpcpb.TrxStatByHourResponse,error) {
	req := &grpcpb.TrxStatByHourRequest{Hours:hours}
	return rpcclient.GetRpc().TrxStatByHour(context.Background(),req)
}

func (w *walletImpl) GetUserTrxListByTime(name string,startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetUserTrxListByTimeRequest{
			Name:prototype.NewAccountName(name),
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastTrx:page.LastOrder.(*grpcpb.TrxInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetUserTrxListByTime(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.TrxList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.TrxList[len(res.TrxList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetPostInfoById(postId uint64) (*grpcpb.GetPostInfoByIdResponse,error) {
	req := &grpcpb.GetPostInfoByIdRequest{
		PostId:postId,
		VoterListLimit:math.MaxUint32,
		ReplyListLimit:math.MaxUint32,
	}
	return rpcclient.GetRpc().GetPostInfoById(context.Background(),req)
}

func (w *walletImpl) GetContractInfo(owner,contract string) (*grpcpb.GetContractInfoResponse,error) {
	req := &grpcpb.GetContractInfoRequest{
		Owner:prototype.NewAccountName(owner),
		ContractName:contract,
		FetchAbi:true,
		FetchCode:true,
	}
	return rpcclient.GetRpc().GetContractInfo(context.Background(),req)
}

func (w *walletImpl) GetBlkIsIrreversibleByTxId(trxId *prototype.Sha256) (*grpcpb.GetBlkIsIrreversibleByTxIdResponse,error) {
	req := &grpcpb.GetBlkIsIrreversibleByTxIdRequest{
		TrxId:trxId,
	}
	return rpcclient.GetRpc().GetBlkIsIrreversibleByTxId(context.Background(),req)
}

func (w *walletImpl) GetAccountListByCreTime(startTime,endTime, pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetAccountListByCreTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastAccount:page.LastOrder.(*grpcpb.AccountInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetAccountListByCreTime(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.List) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.List[len(res.List)-1].Info

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetDailyStats(dapp string,days uint32) (*grpcpb.GetDailyStatsResponse,error) {
	req := &grpcpb.GetDailyStatsRequest{
		Dapp:dapp,
		Days:days,
	}
	return rpcclient.GetRpc().GetDailyStats(context.Background(),req)
}

func (w *walletImpl) GetContractListByTime(startTime,endTime, pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetContractListByTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastContract:page.LastOrder.(*grpcpb.ContractInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetContractListByTime(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.ContractList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.ContractList[len(res.ContractList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetBlockProducerListByVoteCount(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(0)
	end := prototype.NewVest(math.MaxUint64)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetBlockProducerListByVoteCountRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastBlockProducer:page.LastOrder.(*grpcpb.BlockProducerResponse),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetBlockProducerListByVoteCount(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.BlockProducerList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.BlockProducerList[len(res.BlockProducerList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetPostListByVest(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(0)
	end := prototype.NewVest(math.MaxUint64)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetPostListByVestRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastPost:page.LastOrder.(*grpcpb.PostResponse),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetPostListByVest(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.PostList) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.PostList[len(res.PostList)-1]

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) EstimateStamina(transaction *prototype.SignedTransaction) (*grpcpb.EsimateResponse,error) {
	req := &grpcpb.EsimateRequest{
		Transaction:transaction,
	}
	return rpcclient.GetRpc().EstimateStamina(context.Background(),req)
}

func (w *walletImpl) GetNodeNeighbours() (*grpcpb.GetNodeNeighboursResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetNodeNeighbours(context.Background(),req)
}

func (w *walletImpl) GetMyStakers(name string, size uint32) (*grpcpb.GetMyStakerListByNameResponse,error) {
	req := &grpcpb.GetMyStakerListByNameRequest{
		Start:&prototype.StakeRecordReverse{
			To:prototype.NewAccountName(name),
			From:prototype.NewAccountName(""),
		},
		End:&prototype.StakeRecordReverse{
			To:prototype.NewAccountName(name),
			From:prototype.NewAccountName("zzzzzzzzzzzzzzz~"),
		},
		Limit:size,
	}
	return rpcclient.GetRpc().GetMyStakers(context.Background(),req)
}

func (w *walletImpl) GetMyStakes(name string, size uint32) (*grpcpb.GetMyStakeListByNameResponse,error) {
	req := &grpcpb.GetMyStakeListByNameRequest{
		Start:&prototype.StakeRecord{
			From:prototype.NewAccountName(name),
			To:prototype.NewAccountName(""),
		},
		End:&prototype.StakeRecord{
			From:prototype.NewAccountName(name),
			To:prototype.NewAccountName("zzzzzzzzzzzzzzz~"),
		},
		Limit:size,
	}
	return rpcclient.GetRpc().GetMyStakes(context.Background(),req)
}

func (w *walletImpl) GetNodeRunningVersion() (*grpcpb.GetNodeRunningVersionResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetNodeRunningVersion(context.Background(),req)
}

func (w *walletImpl) GetAccountListByVest(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(0)
	end := prototype.NewVest(math.MaxUint64)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		req := &grpcpb.GetAccountListByVestRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastAccount:page.LastOrder.(*grpcpb.AccountInfo),
			Limit:page.Limit,
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetAccountListByVest(context.Background(),req)
		if err != nil {
			return nil,nil,err
		}
		if len(res.List) == 0 {
			return nil,nil,errors.New("empty result")
		}
		lastOrder := res.List[len(res.List)-1].Info

		return res,lastOrder,nil
	})
	return pm,nil
}

func (w *walletImpl) GetBlockProducerByName(name string) (*grpcpb.BlockProducerResponse, error) {
	req := &grpcpb.GetBlockProducerByNameRequest{BpName:prototype.NewAccountName(name)}
	return rpcclient.GetRpc().GetBlockProducerByName(context.Background(),req)
}

func (w *walletImpl) GetAccountByPubKey(pubKey string) (*grpcpb.AccountResponse, error) {
	req := &grpcpb.GetAccountByPubKeyRequest{PublicKey:pubKey}
	return rpcclient.GetRpc().GetAccountByPubKey(context.Background(),req)
}

func (w *walletImpl) GetBlockBFTInfoByNum(blockNum uint64) (*grpcpb.GetBlockBFTInfoByNumResponse, error) {
	req := &grpcpb.GetBlockBFTInfoByNumRequest{BlockNum:blockNum}
	return rpcclient.GetRpc().GetBlockBFTInfoByNum(context.Background(),req)
}

func (w *walletImpl) GetAppTableRecord(table,key string) (*grpcpb.GetAppTableRecordResponse,error) {
	req := &grpcpb.GetAppTableRecordRequest{
		TableName:table,
		Key:key,
	}
	return rpcclient.GetRpc().GetAppTableRecord(context.Background(),req)
}

func (w *walletImpl) GetBlockProducerVoterList(name string) (*grpcpb.GetBlockProducerVoterListResponse,error) {
	req := &grpcpb.GetBlockProducerVoterListRequest{
		BlockProducer:prototype.NewAccountName(name),
		Limit:math.MaxUint32,
		LastVoter:nil,
	}
	return rpcclient.GetRpc().GetBlockProducerVoterList(context.Background(),req)
}

type walletImpl struct {
	accounts map[string]*account.Account
	password string
	fullFileName string
}

func (w *walletImpl) Open(pathToFile, password string) error {
	w.accounts = make(map[string]*account.Account)

	w.fullFileName = pathToFile
	w.password = password

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return w.save()
	} else {
		return w.load()
	}
}

func (w *walletImpl) Add(name, privateKey string) error {
	w.accounts[name] = account.NewAccount(privateKey)
	return w.save()
}

func (w *walletImpl) load() error {

	keyJson, err := ioutil.ReadFile(w.fullFileName)
	if err != nil {
		return err
	}
	var eks EncryptKeyStore
	if err := json.Unmarshal(keyJson, &eks); err != nil {
		return err
	}

	key := []byte(w.password)
	iv, err := base64.StdEncoding.DecodeString(eks.Iv)
	if err != nil {
		return err
	}
	cipher_data, err := base64.StdEncoding.DecodeString(eks.CipherText)
	if err != nil {
		return err
	}
	mac_data, err := base64.StdEncoding.DecodeString(eks.Mac)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(w.password))
	calcMac := mac.Sum(nil)
	if !hmac.Equal(mac_data, calcMac) {
		return errors.New("password incorrect")
	}

	keyStoreData, err := utils.DecryptData(cipher_data, key, iv)
	if err != nil {
		return err
	}

	// gob decode
	var buf bytes.Buffer
	buf.Write(keyStoreData)
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&w.accounts); err != nil {
		return err
	}

	return nil
}

func (w *walletImpl) save() error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(w.accounts)
	if err != nil {
		return err
	}

	// aes encrypted
	cipher_data, iv, err := utils.EncryptData(buf.Bytes(),[]byte(w.password))
	if err != nil {
		return err
	}
	cipher_text := base64.StdEncoding.EncodeToString(cipher_data)
	iv_text := base64.StdEncoding.EncodeToString(iv)
	mac := hmac.New(sha256.New, []byte(w.password))
	calcMac := mac.Sum(nil)
	mac_text := base64.StdEncoding.EncodeToString(calcMac)

	encryptKeyStore := &EncryptKeyStore{
		CipherText: cipher_text,
		Iv:         iv_text,
		Mac:        mac_text,
	}

	// save to file
	return w.seal(encryptKeyStore)
}

func (w *walletImpl) seal(data *EncryptKeyStore) error {

	// I knew there is a problem when user create a pair key but using a name which have been occupied.
	// fixme

	path := filepath.Join(w.fullFileName)
	keyJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, keyJson, 0600)
	return nil
}