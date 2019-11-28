# Contentos golang SDK

Go client library for Contentos blockchain.

## Quick start

### Create a wallet

`Wallet` is the main entry of the library. To create a wallet and connect a certain Contentos node.

```go
wallet := NewKeyStoreWallet("127.0.0.1:8888",utils.Dev)
```
The param utils.Dev  is Chain id you want to use, there are three possible Chain id:

```go
utils.Dev
utils.Test
utils.Main
```

Also a memory wallet, this wallet keep all data in memory without a kehystore file:

```go
w := NewMemWallet("127.0.0.1:8888", utils.Dev)
```

### Open a keystore

```go
file := "/data/test.key"
password := "123"
if err := w.Open(file,password); err != nil {
    return err
}
```

A keystore is just a normal local file that stores all your accounts. Its contents are encrypted for better security, so we need to specify a password for a keystore. In a real app, password should be asked everytime a keystore being opened, and never be stored.

If you pass a non-existent file to `Open()`, a new empty keystore file will be created.

### Import accounts

Once `Open()` is called, you can import your Contentos accounts.

#### Import private key

```go
wallet.Add("yourname","3diUftkv1rsSn45bTNBZgtaYbSstX9eHZfz3WGoX7r7UBsFgLV")
```

If you have multiple accounts, just call `Add()` repeatly to import them all. Imported accounts are permanently stored in the keystore file, you don't have to import them again next time the keystore is opened. 

You can also browse your accounts, query for private keys or remove accounts, Remove function also update keystore file.

```go
accounts := w2.GetAllAccounts()
for k,v := range accounts {
    fmt.Println("name:",k," privateKey:",v)
}
wallet.Remove("sdktest");
```

### Send transactions

```go
acct := "youraccount"
wallet.Add(acct,"3diUftkv1rsSn45bTNBZgtaYbSstX9eHZfz3WGoX7r7UBsFgLV")

benificiary := make(map[string]int)
benificiary[acct] = 2
tags := []string{"1","2"}

res ,err := wallet.Account(acct).Post(acc,"test sdk","sdk",tags,benificiary)
if err != nil {
    return err
}
fmt.Println(res.Invoice)
```

### Query

Contentos provides with rich information of the blockchain. All of these can be retrieved by `Wallet`'s query methods.

```go
// get account information
res,err := wallet.GetAccountByName("sdktest")
if err != nil {
    return err
}
fmt.Println(res)
```

Unlike sending transactions, wallet don't need a private key to make queries.

### List Query

List Query return a PageManager for easy iterate list 

```go
pm,err := wallet.GetAccountListByBalance(1000,1,5)
if err != nil {
	return err
}

for {
    v,err := pm.Next()
        if err != nil {
	    return err
	}

    // different query need specific type cast
    list := v.(*grpcpb.GetAccountListResponse)
	if len(list.List) == 0 {
	    return errors.New("no more results")
	}
	for _,l := range list.List {
	    fmt.Println(l)
	}
}
```

### Close a wallet

When a wallet is no longer needed, don't forget to close it, this will release underlying memory. 

```go
wallet.Close();
```

