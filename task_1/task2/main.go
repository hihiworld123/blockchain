package main

import (
	"blockchain/task_1/task2/abi/counter"
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var client *ethclient.Client
var apiUrl string
var privatekey string

func main() {
	//部署合约
	contractAddress := deploy()
	fmt.Println("Contract Address:", contractAddress)

	//写入智能合约
	setValue(contractAddress)
	//等待区块交易执行完成
	waitForReceipt(contractAddress)
	//查询值，等待执行完成后才能查询到值
	getValue("0x65Bb7CB891AA12688A1ec8F25a237b907Dbb4603")
}

func getValue(contractAddress string) {

	address := common.HexToAddress(contractAddress)
	instance, err := counter.NewCounter(address, client)
	if err != nil {
		log.Fatal(err)
	}
	x, err := instance.X(nil)
	if err != nil {
		log.Fatal("getValue err: ", err)
	}
	fmt.Println("Counter x:", x)
}

func waitForReceipt(contractAddress string) (*types.Receipt, error) {
	txHash := common.HexToHash(contractAddress)
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			return nil, err
		}
		if receipt.Status == types.ReceiptStatusSuccessful {
			return receipt, nil
		}
		time.Sleep(time.Second)
	}
}

// 写入智能合约
func setValue(contractAddress string) {

	address := common.HexToAddress(contractAddress)
	instance, err := counter.NewCounter(address, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("contract is loaded")

	privateKey, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	//设置值
	tx, err := instance.IncBy(auth, big.NewInt(2))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tx Hash: %s\n", tx.Hash().Hex())

}

// 部署合约
func deploy() string {
	privateKey, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := counter.DeployCounter(auth, client)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(address.Hex())
	fmt.Println(tx.Hash().Hex())

	return address.Hex()
}

func init() {
	// 加载.env文件中的环境变量到当前环境中
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file, err: ", err)
	}

	//获取配置信息
	apiUrl = os.Getenv("API_URL")
	privatekey = os.Getenv("PRIVATE_KEY")

	client, err = ethclient.Dial(apiUrl)
	if err != nil {
		log.Fatal("err: ", err)
	}
	log.Println("Connected to Ethereum client", client)
}
