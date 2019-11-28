package wallet

import (
	"github.com/coschain/cos-sdk-go/account"
	"github.com/coschain/cos-sdk-go/rpcclient"
	"github.com/coschain/cos-sdk-go/utils"
)

type MemWallet struct {
	BaseWallet
}

func NewMemWallet(ip string, chainId utils.ChainId) *MemWallet {
	if err := rpcclient.ConnectRpc(ip); err != nil {
		return nil
	}
	w := &MemWallet{}
	w.accounts = make(map[string]*account.Account)
	w.chainId = chainId
	return w
}

func (w *MemWallet) Close() {
	w.accounts = nil
}

func (w *MemWallet) Add(name, privateKey string) {
	w.accounts[name] = account.NewAccount(privateKey, func() utils.ChainId {
		return w.chainId
	})
}

func (w *MemWallet) Remove(name string) {
	delete(w.accounts,name)
}
