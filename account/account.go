package account

import (
	"context"
	"fmt"
	"github.com/coschain/contentos-go/common"
	"github.com/coschain/contentos-go/common/constants"
	"github.com/coschain/contentos-go/prototype"
	"github.com/coschain/contentos-go/rpc/pb"
	"github.com/coschain/cos-sdk-go/rpcclient"
	"github.com/coschain/cos-sdk-go/utils"
	"github.com/kataras/go-errors"
)

type GetChainId func() utils.ChainId

type Account struct {
	Name string
	PrivateKey string
	GetChainIdCallBack GetChainId
}

func NewAccount(name, privateKey string, callBack GetChainId) *Account {
	return &Account{
		Name:name,
		PrivateKey: privateKey,
		GetChainIdCallBack:callBack,
	}
}

func (a *Account) CreateAccount(fee uint64, newAccountName, pubKeyStr, meta string) (*grpcpb.BroadcastTrxResponse,error) {
	pubKey, _ := prototype.PublicKeyFromWIF(pubKeyStr)
	acOp := &prototype.AccountCreateOperation{
		Creator:prototype.NewAccountName(a.Name),
		Fee:prototype.NewCoin(fee),
		NewAccountName:prototype.NewAccountName(newAccountName),
		PubKey:pubKey,
		JsonMetadata:meta,
	}

	return a.broadcastTrx(a.PrivateKey,acOp)
}

func (a *Account) BpRegist(owner, bpUrl, bpDesc, pubKeyStr string, fee, proposedStaminaFree, tpsExpected, bpEpochDuration, ticketPrice, bpPerTicketWeight uint64, bpTopN uint32) (*grpcpb.BroadcastTrxResponse, error) {
	pubKey, _ := prototype.PublicKeyFromWIF(pubKeyStr)
	bpRegistOp := &prototype.BpRegisterOperation{
		Owner:           &prototype.AccountName{Value: owner},
		Url:             bpUrl,
		Desc:            bpDesc,
		BlockSigningKey: pubKey,
		Props: &prototype.ChainProperties{
			AccountCreationFee: prototype.NewCoin(fee),
			StaminaFree:        proposedStaminaFree,
			TpsExpected:        tpsExpected,
			EpochDuration:      bpEpochDuration,
			TopNAcquireFreeToken: bpTopN,
			PerTicketPrice:     prototype.NewCoin(ticketPrice),
			PerTicketWeight:    bpPerTicketWeight,
		},
	}
	return a.broadcastTrx(a.PrivateKey,bpRegistOp)
}

func (a *Account) BpEnable(name string, cancel bool) (*grpcpb.BroadcastTrxResponse, error) {
	bpEnableOp := &prototype.BpEnableOperation{
		Owner:      &prototype.AccountName{Value: name},
		Cancel:     cancel,
	}
	return a.broadcastTrx(a.PrivateKey,bpEnableOp)
}

func (a *Account) BpVote(bp string, cancel bool) (*grpcpb.BroadcastTrxResponse, error) {
	bpVoteOp := &prototype.BpVoteOperation{
		Voter:   &prototype.AccountName{Value: a.Name},
		BlockProducer: &prototype.AccountName{Value: bp},
		Cancel:  cancel,
	}
	return a.broadcastTrx(a.PrivateKey,bpVoteOp)
}

func (a *Account) Post(title,content string,tags []string, postBeneficiaryRoute map[string]int) (*grpcpb.BroadcastTrxResponse, error) {
	beneficiaries := []*prototype.BeneficiaryRouteType{}
	accumulateWeight := 0
	for k, v := range postBeneficiaryRoute {
		if v < 0 {
			return nil,errors.New("weight should greater than zero")
		}

		if v > constants.PERCENT {
			return nil,errors.New("either beneficiary route should not greater than 100%")
		}

		if accumulateWeight > constants.PERCENT {
			return nil,errors.New("accumulated weight should not greater than 100%")
		}

		accumulateWeight += v
		route := &prototype.BeneficiaryRouteType{
			Name:   &prototype.AccountName{Value: k},
			Weight: uint32(v),
		}

		beneficiaries = append(beneficiaries, route)
	}

	uuid := utils.GenerateUUID(a.Name + title)
	postOp := &prototype.PostOperation{
		Uuid:          uuid,
		Owner:         &prototype.AccountName{Value: a.Name},
		Title:         title,
		Content:       content,
		Tags:          tags,
		Beneficiaries: beneficiaries,
	}
	return a.broadcastTrx(a.PrivateKey,postOp)
}

func (a *Account) Reply(content string, postId uint64, replyBeneficiaryRoute map[string]int) (*grpcpb.BroadcastTrxResponse, error) {
	beneficiaries := []*prototype.BeneficiaryRouteType{}
	accumulateWeight := 0
	for k, v := range replyBeneficiaryRoute {
		if v < 0 {
			return nil,errors.New("weight should greater than zero")
		}

		if v > constants.PERCENT {
			return nil,errors.New("either beneficiary route should not greater than 100%")
		}

		if accumulateWeight > constants.PERCENT {
			return nil,errors.New("accumulated weight should not greater than 100%")
		}

		accumulateWeight += v
		route := &prototype.BeneficiaryRouteType{
			Name:   &prototype.AccountName{Value: k},
			Weight: uint32(v),
		}
		beneficiaries = append(beneficiaries, route)
	}
	uuid := utils.GenerateUUID(a.Name)
	replyOp := &prototype.ReplyOperation{
		Uuid:          uuid,
		Owner:         &prototype.AccountName{Value: a.Name},
		Content:       content,
		ParentUuid:    postId,
		Beneficiaries: beneficiaries,
	}
	return a.broadcastTrx(a.PrivateKey,replyOp)
}

func (a *Account) Follow(following string, cancel bool) (*grpcpb.BroadcastTrxResponse, error) {
	followOp := &prototype.FollowOperation{
		Account:  &prototype.AccountName{Value: a.Name},
		FAccount: &prototype.AccountName{Value: following},
		Cancel:   cancel,
	}
	return a.broadcastTrx(a.PrivateKey,followOp)
}

func (a *Account) Vote(idx uint64) (*grpcpb.BroadcastTrxResponse, error) {
	voterOp := &prototype.VoteOperation{
		Voter: &prototype.AccountName{Value: a.Name},
		Idx:   idx,
	}
	return a.broadcastTrx(a.PrivateKey,voterOp)
}

func (a *Account) Transfer(to string,amount uint64, memo string) (*grpcpb.BroadcastTrxResponse, error) {
	transferOp := &prototype.TransferOperation{
		From:   &prototype.AccountName{Value: a.Name},
		To:     &prototype.AccountName{Value: to},
		Amount: prototype.NewCoin(amount),
		Memo:   memo,
	}
	return a.broadcastTrx(a.PrivateKey,transferOp)
}

func (a *Account) ContractDeploy(cname string, abi,code []byte, upgradeable bool, contractUrl,contractDesc string ) (*grpcpb.BroadcastTrxResponse, error) {
	var (
		compressedCode, compressedAbi []byte
		err error
	)
	if compressedCode, err = common.Compress(code); err != nil {
		return nil,errors.New(fmt.Sprintf("code compression failed: %s", err.Error()))
	}
	if compressedAbi, err = common.Compress(abi); err != nil {
		return nil,errors.New(fmt.Sprintf("abi compression failed: %s", err.Error()))
	}

	contractDeployOp := &prototype.ContractDeployOperation{
		Owner:    &prototype.AccountName{Value: a.Name},
		Contract: cname,
		Abi:      compressedAbi,
		Code:     compressedCode,
		Upgradeable:upgradeable,
		Url: contractUrl,
		Describe: contractDesc,
	}
	return a.broadcastTrx(a.PrivateKey,contractDeployOp)
}

func (a *Account) ContractApply(owner,cname,params,method string,fee uint64) (*grpcpb.BroadcastTrxResponse, error) {
	contractApplyOp := &prototype.ContractApplyOperation{
		Caller:   &prototype.AccountName{Value: a.Name},
		Owner:    &prototype.AccountName{Value: owner},
		Amount:   &prototype.Coin{Value: fee},
		Contract: cname,
		Params:   params,
		Method:	  method,
	}
	return a.broadcastTrx(a.PrivateKey,contractApplyOp)
}

func (a *Account) ConvertVest(amount uint64) (*grpcpb.BroadcastTrxResponse, error) {
	convertVestOp := &prototype.ConvertVestOperation{
		From:   &prototype.AccountName{Value: a.Name},
		Amount: prototype.NewVest(uint64(amount)),
	}
	return a.broadcastTrx(a.PrivateKey,convertVestOp)
}

func (a *Account) Stake(to string, amount uint64) (*grpcpb.BroadcastTrxResponse, error) {
	stakeOp := &prototype.StakeOperation{
		From:   &prototype.AccountName{Value: a.Name},
		To:   &prototype.AccountName{Value: to},
		Amount:    prototype.NewCoin(amount),
	}
	return a.broadcastTrx(a.PrivateKey,stakeOp)
}

func (a *Account) UnStake(debtor string, amount uint64) (*grpcpb.BroadcastTrxResponse, error) {
	unStakeOp := &prototype.UnStakeOperation{
		Creditor:   &prototype.AccountName{Value: a.Name},
		Debtor:   &prototype.AccountName{Value: debtor},
		Amount:    prototype.NewCoin(amount),
	}
	return a.broadcastTrx(a.PrivateKey,unStakeOp)
}

func (a *Account) BpUpdate(name string,bpUpdateStaminaFree,bpUpdateTpsExpected,bpUpdateEpochDuration,bpUpdatePerTicketWeight,bpUpdateCreateAccountFee,bpUpdatePerTicketPrice uint64,bpUpdateTopN uint32) (*grpcpb.BroadcastTrxResponse, error) {
	props := &prototype.ChainProperties{}
	props.StaminaFree = bpUpdateStaminaFree
	props.TpsExpected = bpUpdateTpsExpected
	props.PerTicketPrice = prototype.NewCoin(bpUpdatePerTicketPrice)
	props.AccountCreationFee = prototype.NewCoin(bpUpdateCreateAccountFee)
	props.TopNAcquireFreeToken = bpUpdateTopN
	props.EpochDuration = bpUpdateEpochDuration
	props.PerTicketWeight = bpUpdatePerTicketWeight

	bpUpdateOp := &prototype.BpUpdateOperation{
		Owner:                 &prototype.AccountName{Value: name},
		Props:                 props,
	}
	return a.broadcastTrx(a.PrivateKey,bpUpdateOp)
}

func (a *Account) AccountUpdate(pubKeyStr string) (*grpcpb.BroadcastTrxResponse, error) {
	pubKey, err := prototype.PublicKeyFromWIF(pubKeyStr)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	accountUpdateOp := &prototype.AccountUpdateOperation{
		Owner:         &prototype.AccountName{Value: a.Name},
		PubKey:        pubKey,
	}
	return a.broadcastTrx(a.PrivateKey,accountUpdateOp)
}

func (a *Account) AcquireTicket(name string, count uint64) (*grpcpb.BroadcastTrxResponse, error) {
	acquireTicketOp := &prototype.AcquireTicketOperation{
		Account: &prototype.AccountName{Value:name},
		Count: count,
	}
	return a.broadcastTrx(a.PrivateKey,acquireTicketOp)
}

func (a *Account) VoteByTicket(name string,postId,count uint64) (*grpcpb.BroadcastTrxResponse, error) {
	voteByTicketOp := &prototype.VoteByTicketOperation{
		Account: &prototype.AccountName{Value:name},
		Idx: postId,
		Count: count,
	}
	return a.broadcastTrx(a.PrivateKey,voteByTicketOp)
}


func (a *Account) TransferToVest(to string, amount uint64, memo string) (*grpcpb.BroadcastTrxResponse, error) {
	transferToVestOp := &prototype.TransferToVestOperation{
		From:prototype.NewAccountName(a.Name),
		To:prototype.NewAccountName(to),
		Amount:prototype.NewCoin(amount),
		Memo:memo,
	}
	return a.broadcastTrx(a.PrivateKey,transferToVestOp)
}

func (a *Account) DelegateVest(to string, amount uint64, expiration uint64) (*grpcpb.BroadcastTrxResponse, error) {
	delegateVestOp := &prototype.DelegateVestOperation{
		From:prototype.NewAccountName(a.Name),
		To:prototype.NewAccountName(to),
		Amount:prototype.NewVest(amount),
		Expiration:expiration,
	}
	return a.broadcastTrx(a.PrivateKey,delegateVestOp)
}

func (a *Account) UnDelegateVest(orderId uint64) (*grpcpb.BroadcastTrxResponse, error) {
	unDelegateVestOp := &prototype.UnDelegateVestOperation{
		Account:prototype.NewAccountName(a.Name),
		OrderId:orderId,
	}
	return a.broadcastTrx(a.PrivateKey,unDelegateVestOp)
}

func (a *Account) broadcastTrx(privateKey string, op ...interface{}) (*grpcpb.BroadcastTrxResponse,error) {
	signTx, err := utils.GenerateSignedTxAndValidate(rpcclient.GetRpc(), privateKey, string(a.GetChainIdCallBack()),op...)
	if err != nil {
		return nil,err
	}
	req := &grpcpb.BroadcastTrxRequest{Transaction: signTx}
	res, err := rpcclient.GetRpc().BroadcastTrx(context.Background(),req)
	return res,err
}
