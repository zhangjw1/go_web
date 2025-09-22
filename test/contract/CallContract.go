package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	store "go-web-starter/contract"
	"log"
	"math/big"
	"strings"
)

func callContract() {
	// 创建一个Ethereum客户端
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/efe70fa30ffd4eb08d17e189753d80b9")
	if err != nil {
		log.Fatal(err)
	}
	// 创建一个Store合约实例
	storeContract, err := store.NewStore(common.HexToAddress("0x330d9507eBC56Cc8d63273d1BcaaedFA4F679619"), client)
	if err != nil {
		log.Fatal(err)
	}
	//获取私钥
	privateKey, err := crypto.HexToECDSA("fc56c167e6b531c6752ee8a5458148753a08a1ec3944b46af7c578e073424e97")
	if err != nil {
		log.Fatal(err)
	}

	// 设置键值对
	var key [32]byte
	var value [32]byte
	copy(key[:], []byte("key"))
	copy(value[:], []byte("value"))

	// 初始化交易opt实例
	opt, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(11155111))
	if err != nil {
		log.Fatal(err)
	}
	//调用合约方法SetItem
	tx, err := storeContract.SetItem(opt, key, value)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("tx sent: %s\n", tx.Hash().Hex())

	//查询合约中的数据并验证
	callOpt := bind.CallOpts{Context: context.Background()}
	// 调用合约GetItem方法
	valueInContract, err := storeContract.Items(&callOpt, key)
	if err != nil {
		log.Fatal(err)
	}
	valueStr := string(bytes.TrimRight(valueInContract[:], "\x00"))
	fmt.Printf("value in contract: %s\n", valueStr)
	fmt.Println("is value saving in contract equals to origin value:", valueInContract == value)

}

const (
	// store合约的字节码
	contractAddress = "0x29e278048a4c654Dc0b77479f598D2b6810B27F7"
)

func callContractByABI() {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/efe70fa30ffd4eb08d17e189753d80b9")
	if err != nil {
		log.Fatal(err)
	}

	//获取私钥
	privateKey, err := crypto.HexToECDSA("fc56c167e6b531c6752ee8a5458148753a08a1ec3944b46af7c578e073424e97")
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	//公钥的椭圆圆锥曲线
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	//公钥对用的钱包 地址
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	//获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		log.Fatal(err)
	}
	//获取gasPrice
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	//组装交易数据
	contractABI, err := abi.JSON(strings.NewReader(`[{"inputs":[{"internalType":"string","name":"_version","type":"string"}],"stateMutability":"nonpayable","type":"constructor"},{"anonymous":false,"inputs":[{"indexed":false,"internalType":"bytes32","name":"key","type":"bytes32"},{"indexed":false,"internalType":"bytes32","name":"value","type":"bytes32"}],"name":"ItemSet","type":"event"},{"inputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"name":"items","outputs":[{"internalType":"bytes32","name":"","type":"bytes32"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"bytes32","name":"key","type":"bytes32"},{"internalType":"bytes32","name":"value","type":"bytes32"}],"name":"setItem","outputs":[],"stateMutability":"nonpayable","type":"function"},{"inputs":[],"name":"version","outputs":[{"internalType":"string","name":"","type":"string"}],"stateMutability":"view","type":"function"}]`))
	methodName := "setItem"
	var key [32]byte
	var value [32]byte
	copy(key[:], []byte("demo_save_key_use_abi"))
	copy(value[:], []byte("demo_save_value_use_abi_11111"))
	//Pack用于将函数调用和参数打包成以太坊智能合约可以理解的二进制格式
	input, err := contractABI.Pack(methodName, key, value)
	if err != nil {
		log.Fatal(err)
	}
	cahinID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress(contractAddress)
	//创建交易对象
	tx := types.NewTransaction(nonce, toAddress, big.NewInt(0), 300000, gasPrice, input)
	//使用私钥对交易进行签名
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(cahinID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	//将签名的交易发送到以太坊网络
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tx sent: %s\n", signedTx.Hash())
	//等待交易执行完成
	_, err = waitForReceipt(client, signedTx.Hash())
	if err != nil {
		log.Fatal(err)
	}

	// 1. 使用ABI Pack方法编码函数调用（与setItem相同）
	callInput, err := contractABI.Pack("items", key)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 构造调用消息
	callMsg := ethereum.CallMsg{
		To:   &toAddress,
		Data: callInput,
	}
	// 3. 直接执行调用并获取结果，不需要改变链上数据，不需要像调用setItems那样复杂
	result, err := client.CallContract(context.Background(), callMsg, nil)
	if err != nil {
		log.Fatal(err)
	}

	var unpacked [32]byte
	contractABI.UnpackIntoInterface(&unpacked, "items", result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("is value saving in contract equals to origin value:", unpacked == value)
}
