package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

var client *ethclient.Client
var apiUrl string
var toaddress string
var privatekey string

func main() {
	queryBlock()
	transfer()
}

// 查询区块
func queryBlock() {
	//查询区块信息
	blockNumber := big.NewInt(9781341)
	block, err2 := client.BlockByNumber(context.Background(), blockNumber)
	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println("block hash:", block.Hash().String())
	fmt.Println("block timestamp:", block.Time())
	fmt.Println("block transactions len:", block.Transactions().Len())
}

// 发送交易
func transfer() {
	privateKey, err := crypto.HexToECDSA(privatekey)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("transfer nonce", nonce)

	value := big.NewInt(10000000000000000) //0.01 eth
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	toAddress := common.HexToAddress(toaddress)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("transfer signedTx", signedTx.Hash().Hex())

}

func init() {
	// 加载.env文件中的环境变量到当前环境中
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file, err: ", err)
	}

	//获取配置信息
	apiUrl = os.Getenv("API_URL")
	toaddress = os.Getenv("TO_ADDRESS")
	privatekey = os.Getenv("PRIVATE_KEY")

	client, err = ethclient.Dial(apiUrl)
	if err != nil {
		log.Fatal("err: ", err)
	}
	log.Println("Connected to Ethereum client", client)
}
