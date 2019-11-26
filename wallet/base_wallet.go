package wallet

import (
	"github.com/coschain/contentos-go/prototype"
	"github.com/coschain/contentos-go/rpc/pb"
	"github.com/coschain/cos-sdk-go/account"
	"github.com/coschain/cos-sdk-go/rpcclient"
	"context"
	"math"
	"errors"
)

type BaseWallet struct {
	accounts map[string]*account.Account
}

func (w *BaseWallet) Account(name string) *account.Account {
	return w.accounts[name]
}

func (w *BaseWallet) GetAllAccounts() map[string]*account.Account {
	return w.accounts
}

func (w *BaseWallet) QueryTableContent(owner,contract,table,field string, count uint32, reverse bool) (*grpcpb.TableContentResponse,error) {
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

func (w *BaseWallet) GetAccountByName(name string) (*grpcpb.AccountResponse,error) {
	req := &grpcpb.GetAccountByNameRequest{AccountName: &prototype.AccountName{Value: name}}
	return rpcclient.GetRpc().GetAccountByName(context.Background(), req)
}

func (w *BaseWallet) GetFollowerListByName(name string, pageSize uint32) (*PageManager,error) {

	start := &prototype.FollowerCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:math.MaxUint32},
		Follower:prototype.NewAccountName(""),
	}
	end := &prototype.FollowerCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:0},
		Follower:prototype.NewAccountName(""),
	}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *prototype.FollowerCreatedOrder
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*prototype.FollowerCreatedOrder)
		}
		req := &grpcpb.GetFollowerListByNameRequest{
			Start:page.Start.(*prototype.FollowerCreatedOrder),
			End:page.End.(*prototype.FollowerCreatedOrder),
			Limit:page.Limit,
			LastOrder:last,
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
	},func(lastOrder interface{}) interface{} {
		return lastOrder
	})
	return pm,nil
}

func (w *BaseWallet) GetFollowingListByName(name string, pageSize uint32) (*PageManager,error) {
	start := &prototype.FollowingCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:math.MaxUint32},
		Following:prototype.NewAccountName(""),
	}
	end := &prototype.FollowingCreatedOrder{
		Account:prototype.NewAccountName(name),
		CreatedTime:&prototype.TimePointSec{UtcSeconds:0},
		Following:prototype.NewAccountName(""),
	}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *prototype.FollowingCreatedOrder
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*prototype.FollowingCreatedOrder)
		}
		req := &grpcpb.GetFollowingListByNameRequest{
			Start:page.Start.(*prototype.FollowingCreatedOrder),
			End:page.End.(*prototype.FollowingCreatedOrder),
			Limit:page.Limit,
			LastOrder:last,
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
	},func(lastOrder interface{}) interface{} {
		return lastOrder
	})
	return pm,nil
}

func (w *BaseWallet) GetFollowCountByName(name string) (*grpcpb.GetFollowCountByNameResponse,error) {
	req := &grpcpb.GetFollowCountByNameRequest{
		AccountName:prototype.NewAccountName(name),
	}
	return rpcclient.GetRpc().GetFollowCountByName(context.Background(),req)
}

func (w *BaseWallet) GetBlockProducerList(size uint32) (*grpcpb.GetBlockProducerListResponse,error) {
	req := &grpcpb.GetBlockProducerListRequest{
		Start:prototype.NewAccountName(""),
		Limit:size,
	}
	return rpcclient.GetRpc().GetBlockProducerList(context.Background(),req)
}

func (w *BaseWallet) GetPostListByCreated(startTime uint32, endTime uint32, limit uint32) (*grpcpb.GetPostListByCreatedResponse,error) {
	req := &grpcpb.GetPostListByCreatedRequest{
		Start:&prototype.PostCreatedOrder{Created:prototype.NewTimePointSec(startTime)},
		End:&prototype.PostCreatedOrder{Created:prototype.NewTimePointSec(endTime)},
		Limit:limit,
	}
	return rpcclient.GetRpc().GetPostListByCreated(context.Background(),req)
}

func (w *BaseWallet) GetReplyListByPostId(postid uint64, startTime uint32, endTime uint32, limit uint32) (*grpcpb.GetReplyListByPostIdResponse,error) {
	req := &grpcpb.GetReplyListByPostIdRequest{
		Start:&prototype.ReplyCreatedOrder{ParentId:postid,Created:prototype.NewTimePointSec(math.MaxUint32)},
		End:&prototype.ReplyCreatedOrder{ParentId:postid,Created:prototype.NewTimePointSec(0)},
		Limit:limit,
	}
	return rpcclient.GetRpc().GetReplyListByPostId(context.Background(),req)
}

func (w *BaseWallet) GetBlockTransactionsByNum(blockNum uint32) (*grpcpb.GetBlockTransactionsByNumResponse,error) {
	req := &grpcpb.GetBlockTransactionsByNumRequest{BlockNum:blockNum}
	return rpcclient.GetRpc().GetBlockTransactionsByNum(context.Background(),req)
}

func (w *BaseWallet) GetChainState() (*grpcpb.GetChainStateResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetChainState(context.Background(),req)
}

func (w *BaseWallet) GetBlockList(start, end uint64, limit uint32) (*grpcpb.GetBlockListResponse,error) {
	req := &grpcpb.GetBlockListRequest{
		Start:start,
		End:end,
		Limit:limit,
	}
	return rpcclient.GetRpc().GetBlockList(context.Background(),req)
}

func (w *BaseWallet) GetSignedBlock(blockNum uint64) (*grpcpb.GetSignedBlockResponse,error) {
	req := &grpcpb.GetSignedBlockRequest{Start:blockNum}
	return rpcclient.GetRpc().GetSignedBlock(context.Background(),req)
}

func (w *BaseWallet) GetAccountListByBalance(startCoin,endCoin uint64, pageSize uint32) (*PageManager,error) {
	start := prototype.NewCoin(startCoin)
	end := prototype.NewCoin(endCoin)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.AccountInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.AccountInfo)
		}
		req := &grpcpb.GetAccountListByBalanceRequest{
			Start:page.Start.(*prototype.Coin),
			End:page.End.(*prototype.Coin),
			LastAccount:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.AccountInfo)
		return v.Coin
	})
	return pm,nil
}

func (w *BaseWallet) GetDailyTotalTrxInfo(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.DailyTotalTrx
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.DailyTotalTrx)
		}
		req := &grpcpb.GetDailyTotalTrxRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastInfo:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.DailyTotalTrx)
		return v.Date
	})
	return pm,nil
}

func (w *BaseWallet) GetTrxInfoById(trxId *prototype.Sha256) (*grpcpb.GetTrxInfoByIdResponse,error) {
	req := &grpcpb.GetTrxInfoByIdRequest{TrxId:trxId}
	return rpcclient.GetRpc().GetTrxInfoById(context.Background(),req)
}

func (w *BaseWallet) GetTrxListByTime(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.TrxInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.TrxInfo)
		}
		req := &grpcpb.GetTrxListByTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastInfo:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.TrxInfo)
		return v.BlockTime
	})
	return pm,nil
}

func (w *BaseWallet) GetPostListByCreateTime(startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.PostResponse
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.PostResponse)
		}
		req := &grpcpb.GetPostListByCreateTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastPost:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.PostResponse)
		return v.Created
	})
	return pm,nil
}

func (w *BaseWallet) GetPostListByName(name string,pageSize uint32) (*PageManager,error) {
	start := &prototype.UserPostCreateOrder{Author:prototype.NewAccountName(name),Create:prototype.NewTimePointSec(math.MaxUint32)}
	end := &prototype.UserPostCreateOrder{Author:prototype.NewAccountName(name),Create:prototype.NewTimePointSec(0)}

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.PostResponse
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.PostResponse)
		}
		req := &grpcpb.GetPostListByNameRequest{
			Start:page.Start.(*prototype.UserPostCreateOrder),
			End:page.End.(*prototype.UserPostCreateOrder),
			LastPost:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.PostResponse)
		return &prototype.UserPostCreateOrder{Author:prototype.NewAccountName(name),Create:v.Created}
	})
	return pm,nil
}

func (w *BaseWallet) TrxStatByHour(hours uint32) (*grpcpb.TrxStatByHourResponse,error) {
	req := &grpcpb.TrxStatByHourRequest{Hours:hours}
	return rpcclient.GetRpc().TrxStatByHour(context.Background(),req)
}

func (w *BaseWallet) GetUserTrxListByTime(name string,startTime,endTime,pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.TrxInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.TrxInfo)
		}
		req := &grpcpb.GetUserTrxListByTimeRequest{
			Name:prototype.NewAccountName(name),
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastTrx:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.TrxInfo)
		return v.BlockTime
	})
	return pm,nil
}

func (w *BaseWallet) GetPostInfoById(postId uint64) (*grpcpb.GetPostInfoByIdResponse,error) {
	req := &grpcpb.GetPostInfoByIdRequest{
		PostId:postId,
		VoterListLimit:math.MaxUint32,
		ReplyListLimit:math.MaxUint32,
	}
	return rpcclient.GetRpc().GetPostInfoById(context.Background(),req)
}

func (w *BaseWallet) GetContractInfo(owner,contract string) (*grpcpb.GetContractInfoResponse,error) {
	req := &grpcpb.GetContractInfoRequest{
		Owner:prototype.NewAccountName(owner),
		ContractName:contract,
		FetchAbi:true,
		FetchCode:true,
	}
	return rpcclient.GetRpc().GetContractInfo(context.Background(),req)
}

func (w *BaseWallet) GetBlkIsIrreversibleByTxId(trxId *prototype.Sha256) (*grpcpb.GetBlkIsIrreversibleByTxIdResponse,error) {
	req := &grpcpb.GetBlkIsIrreversibleByTxIdRequest{
		TrxId:trxId,
	}
	return rpcclient.GetRpc().GetBlkIsIrreversibleByTxId(context.Background(),req)
}

func (w *BaseWallet) GetAccountListByCreTime(startTime,endTime, pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.AccountInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.AccountInfo)
		}
		req := &grpcpb.GetAccountListByCreTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastAccount:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.AccountInfo)
		return v.CreatedTime
	})
	return pm,nil
}

func (w *BaseWallet) GetDailyStats(dapp string,days uint32) (*grpcpb.GetDailyStatsResponse,error) {
	req := &grpcpb.GetDailyStatsRequest{
		Dapp:dapp,
		Days:days,
	}
	return rpcclient.GetRpc().GetDailyStats(context.Background(),req)
}

func (w *BaseWallet) GetContractListByTime(startTime,endTime, pageSize uint32) (*PageManager,error) {
	start := prototype.NewTimePointSec(startTime)
	end := prototype.NewTimePointSec(endTime)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.ContractInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.ContractInfo)
		}
		req := &grpcpb.GetContractListByTimeRequest{
			Start:page.Start.(*prototype.TimePointSec),
			End:page.End.(*prototype.TimePointSec),
			LastContract:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.ContractInfo)
		return v.CreateTime
	})
	return pm,nil
}

func (w *BaseWallet) GetBlockProducerListByVoteCount(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(math.MaxUint64)
	end := prototype.NewVest(0)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.BlockProducerResponse
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.BlockProducerResponse)
		}
		req := &grpcpb.GetBlockProducerListByVoteCountRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastBlockProducer:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.BlockProducerResponse)
		return v.BpVest.VoteVest
	})
	return pm,nil
}

func (w *BaseWallet) GetPostListByVest(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(math.MaxUint64)
	end := prototype.NewVest(0)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.PostResponse
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.PostResponse)
		}
		req := &grpcpb.GetPostListByVestRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastPost:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.PostResponse)
		return v.Rewards // is this atrribute ?
	})
	return pm,nil
}

func (w *BaseWallet) EstimateStamina(transaction *prototype.SignedTransaction) (*grpcpb.EsimateResponse,error) {
	req := &grpcpb.EsimateRequest{
		Transaction:transaction,
	}
	return rpcclient.GetRpc().EstimateStamina(context.Background(),req)
}

func (w *BaseWallet) GetNodeNeighbours() (*grpcpb.GetNodeNeighboursResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetNodeNeighbours(context.Background(),req)
}

func (w *BaseWallet) GetMyStakers(name string, size uint32) (*grpcpb.GetMyStakerListByNameResponse,error) {
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

func (w *BaseWallet) GetMyStakes(name string, size uint32) (*grpcpb.GetMyStakeListByNameResponse,error) {
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

func (w *BaseWallet) GetNodeRunningVersion() (*grpcpb.GetNodeRunningVersionResponse,error) {
	req := &grpcpb.NonParamsRequest{}
	return rpcclient.GetRpc().GetNodeRunningVersion(context.Background(),req)
}

func (w *BaseWallet) GetAccountListByVest(pageSize uint32) (*PageManager,error) {
	start := prototype.NewVest(math.MaxUint64)
	end := prototype.NewVest(0)

	pm := NewPageManager(start,end,pageSize,nil,func(page *Page) (interface{},interface{},error) {
		var last *grpcpb.AccountInfo
		if page.LastOrder == nil {
			last = nil
		} else {
			last = page.LastOrder.(*grpcpb.AccountInfo)
		}
		req := &grpcpb.GetAccountListByVestRequest{
			Start:page.Start.(*prototype.Vest),
			End:page.End.(*prototype.Vest),
			LastAccount:last,
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
	},func(lastOrder interface{}) interface{} {
		v := lastOrder.(*grpcpb.AccountInfo)
		return v.Vest
	})
	return pm,nil
}

func (w *BaseWallet) GetBlockProducerByName(name string) (*grpcpb.BlockProducerResponse, error) {
	req := &grpcpb.GetBlockProducerByNameRequest{BpName:prototype.NewAccountName(name)}
	return rpcclient.GetRpc().GetBlockProducerByName(context.Background(),req)
}

func (w *BaseWallet) GetAccountByPubKey(pubKey string) (*grpcpb.AccountResponse, error) {
	req := &grpcpb.GetAccountByPubKeyRequest{PublicKey:pubKey}
	return rpcclient.GetRpc().GetAccountByPubKey(context.Background(),req)
}

func (w *BaseWallet) GetBlockBFTInfoByNum(blockNum uint64) (*grpcpb.GetBlockBFTInfoByNumResponse, error) {
	req := &grpcpb.GetBlockBFTInfoByNumRequest{BlockNum:blockNum}
	return rpcclient.GetRpc().GetBlockBFTInfoByNum(context.Background(),req)
}

func (w *BaseWallet) GetAppTableRecord(table,key string) (*grpcpb.GetAppTableRecordResponse,error) {
	req := &grpcpb.GetAppTableRecordRequest{
		TableName:table,
		Key:key,
	}
	return rpcclient.GetRpc().GetAppTableRecord(context.Background(),req)
}

func (w *BaseWallet) GetBlockProducerVoterList(name string) (*grpcpb.GetBlockProducerVoterListResponse,error) {
	req := &grpcpb.GetBlockProducerVoterListRequest{
		BlockProducer:prototype.NewAccountName(name),
		Limit:math.MaxUint32,
		LastVoter:nil,
	}
	return rpcclient.GetRpc().GetBlockProducerVoterList(context.Background(),req)
}