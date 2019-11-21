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


	pm := NewPageManager(start,end,pageSize,nil,func(manager *PageManager) interface{} {
		// find latest next page
		if manager.CurrentPage() + 1 > manager.PageCount() {
			return nil
		}
		page := manager.pageList[manager.CurrentPage()]
		req := &grpcpb.GetFollowerListByNameRequest{
			Start:page.Start.(*prototype.FollowerCreatedOrder),
			End:page.End.(*prototype.FollowerCreatedOrder),
			Limit:page.Limit,
			LastOrder:page.LastOrder.(*prototype.FollowerCreatedOrder),
		}

		// call rpc
		res,err := rpcclient.GetRpc().GetFollowerListByName(context.Background(),req)
		if err != nil || len(res.FollowerList) == 0 {
			return nil
		}

		// add new page for next query
		lastOrder := res.FollowerList[len(res.FollowerList)-1].CreateOrder
		manager.pageList = append(manager.pageList,&Page{Start:page.End,End:page.End,Limit:page.Limit,LastOrder:lastOrder})
		manager.pageIndex++
		return res
	})
	return pm,nil
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