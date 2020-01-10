package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"github.com/coschain/cos-sdk-go/account"
	"github.com/coschain/cos-sdk-go/rpcclient"
	"github.com/coschain/cos-sdk-go/utils"
	"github.com/kataras/go-errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type EncryptKeyStore struct {
	CipherText string // encrypted privkey
	Iv         string // the iv
	Mac        string // the mac of passphrase
}

type KeyStoreWallet struct {
	BaseWallet

	password string
	fullFileName string
}

func NewKeyStoreWallet(ip string, chainId utils.ChainId) *KeyStoreWallet {
	if err := rpcclient.ConnectRpc(ip); err != nil {
		return nil
	}
	w := &KeyStoreWallet{}
	w.accounts = make(map[string]*account.Account)
	w.chainId = chainId
	return w
}

func (w *KeyStoreWallet) Open(pathToFile, password string) error {

	w.fullFileName = pathToFile
	w.password = password

	if _, err := os.Stat(pathToFile); os.IsNotExist(err) {
		return w.save()
	} else {
		return w.load()
	}
}

func (w *KeyStoreWallet) Close() {
	w.accounts = nil
}

func (w *KeyStoreWallet) Add(name, privateKey string) error {
	w.accounts[name] = account.NewAccount(name, privateKey, func() utils.ChainId {
		return w.chainId
	})
	return w.save()
}

func (w *KeyStoreWallet) AddByMnemonic(name, mnemonic string) error {
	_,pri,err := w.GenerateKeyPairFromMnemonic(mnemonic)
	if err != nil {
		return err
	}
	return w.Add(name,pri)
}

func (w *KeyStoreWallet) Remove(name string) error {
	delete(w.accounts,name)
	return w.save()
}

func (w *KeyStoreWallet) load() error {

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

	// set call back func
	for _,v := range w.accounts {
		v.GetChainIdCallBack = func() utils.ChainId {
			return w.chainId
		}
	}

	return nil
}

func (w *KeyStoreWallet) save() error {
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

func (w *KeyStoreWallet) seal(data *EncryptKeyStore) error {

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