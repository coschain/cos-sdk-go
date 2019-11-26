package wallet

import (
	"github.com/coschain/cos-sdk-go/account"
)

type MemWallet struct {
	BaseWallet
}

func (w *MemWallet) Open(pathToFile, password string) {
	w.accounts = make(map[string]*account.Account)
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
