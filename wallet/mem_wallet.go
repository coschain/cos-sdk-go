package wallet

import (
	"github.com/coschain/cos-sdk-go/account"
	"github.com/coschain/cos-sdk-go/rpcclient"
)

type MemWallet struct {
	BaseWallet
}

func NewMemWallet(ip string) *MemWallet {
	if err := rpcclient.ConnectRpc(ip); err != nil {
		return nil
	}
	w := &MemWallet{}
	w.accounts = make(map[string]*account.Account)
	return w
}

func (w *MemWallet) Close() {
	w.accounts = nil
}

func (w *MemWallet) Add(name, privateKey string) {
	w.accounts[name] = account.NewAccount(privateKey)
}

func (w *MemWallet) Remove(name string) {
	delete(w.accounts,name)
}
