package wallet

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
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

type QueryRpc interface {
	GetAccountByName(name string) error
}

type Wallet interface {
	QueryRpc

	Open(path string, password string) error
	Add(name, privateKey string) error
	Remove(name string) error
	Account(name string) *Account
	Close()
}

type walletImpl struct {
	accounts map[string]string
	password string
	fullFileName string
}

func (w *walletImpl) Open(path, password string) error {
	w.accounts = make(map[string]string)

	w.fullFileName = path
	w.password = password

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return w.save()
	} else {
		return w.load()
	}
}

func (w *walletImpl) Add(name, privateKey string) error {
	w.accounts[name] = privateKey
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