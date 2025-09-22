package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
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
	"runtime/debug"
	"strings"
	"time"
)

func deployByAbi() {
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/efe70fa30ffd4eb08d17e189753d80b9")
	if err != nil {
		log.Fatal(err)
	}

	// privateKey, err := crypto.GenerateKey()
	// privateKeyBytes := crypto.FromECDSA(privateKey)
	// privateKeyHex := hex.EncodeToString(privateKeyBytes)
	// fmt.Println("Private Key:", privateKeyHex)
	privateKey, err := crypto.HexToECDSA("fc56c167e6b531c6752ee8a5458148753a08a1ec3944b46af7c578e073424e97")
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

	chainId, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		log.Fatal(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, instance, err := store.DeployStore(auth, client, input)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("contract address:%v\n", address.Hex())
	fmt.Printf("contract tx hash:%v\n", tx.Hash().Hex())

	_ = instance
}

const (
	// store合约的字节码
	contractBytecode = "608060405234801561000f575f5ffd5b5060405161087838038061087883398181016040528101906100319190610193565b805f908161003f91906103ea565b50506104b9565b5f604051905090565b5f5ffd5b5f5ffd5b5f5ffd5b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6100a58261005f565b810181811067ffffffffffffffff821117156100c4576100c361006f565b5b80604052505050565b5f6100d6610046565b90506100e2828261009c565b919050565b5f67ffffffffffffffff8211156101015761010061006f565b5b61010a8261005f565b9050602081019050919050565b8281835e5f83830152505050565b5f610137610132846100e7565b6100cd565b9050828152602081018484840111156101535761015261005b565b5b61015e848285610117565b509392505050565b5f82601f83011261017a57610179610057565b5b815161018a848260208601610125565b91505092915050565b5f602082840312156101a8576101a761004f565b5b5f82015167ffffffffffffffff8111156101c5576101c4610053565b5b6101d184828501610166565b91505092915050565b5f81519050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061022857607f821691505b60208210810361023b5761023a6101e4565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261029d7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82610262565b6102a78683610262565b95508019841693508086168417925050509392505050565b5f819050919050565b5f819050919050565b5f6102eb6102e66102e1846102bf565b6102c8565b6102bf565b9050919050565b5f819050919050565b610304836102d1565b610318610310826102f2565b84845461026e565b825550505050565b5f5f905090565b61032f610320565b61033a8184846102fb565b505050565b5b8181101561035d576103525f82610327565b600181019050610340565b5050565b601f8211156103a25761037381610241565b61037c84610253565b8101602085101561038b578190505b61039f61039785610253565b83018261033f565b50505b505050565b5f82821c905092915050565b5f6103c25f19846008026103a7565b1980831691505092915050565b5f6103da83836103b3565b9150826002028217905092915050565b6103f3826101da565b67ffffffffffffffff81111561040c5761040b61006f565b5b6104168254610211565b610421828285610361565b5f60209050601f831160018114610452575f8415610440578287015190505b61044a85826103cf565b8655506104b1565b601f19841661046086610241565b5f5b8281101561048757848901518255600182019150602085019450602081019050610462565b868310156104a457848901516104a0601f8916826103b3565b8355505b6001600288020188555050505b505050505050565b6103b2806104c65f395ff3fe608060405234801561000f575f5ffd5b506004361061003f575f3560e01c806348f343f31461004357806354fd4d5014610073578063f56256c714610091575b5f5ffd5b61005d600480360381019061005891906101d7565b6100ad565b60405161006a9190610211565b60405180910390f35b61007b6100c2565b604051610088919061029a565b60405180910390f35b6100ab60048036038101906100a691906102ba565b61014d565b005b6001602052805f5260405f205f915090505481565b5f80546100ce90610325565b80601f01602080910402602001604051908101604052809291908181526020018280546100fa90610325565b80156101455780601f1061011c57610100808354040283529160200191610145565b820191905f5260205f20905b81548152906001019060200180831161012857829003601f168201915b505050505081565b8060015f8481526020019081526020015f20819055507fe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d48282604051610194929190610355565b60405180910390a15050565b5f5ffd5b5f819050919050565b6101b6816101a4565b81146101c0575f5ffd5b50565b5f813590506101d1816101ad565b92915050565b5f602082840312156101ec576101eb6101a0565b5b5f6101f9848285016101c3565b91505092915050565b61020b816101a4565b82525050565b5f6020820190506102245f830184610202565b92915050565b5f81519050919050565b5f82825260208201905092915050565b8281835e5f83830152505050565b5f601f19601f8301169050919050565b5f61026c8261022a565b6102768185610234565b9350610286818560208601610244565b61028f81610252565b840191505092915050565b5f6020820190508181035f8301526102b28184610262565b905092915050565b5f5f604083850312156102d0576102cf6101a0565b5b5f6102dd858286016101c3565b92505060206102ee858286016101c3565b9150509250929050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f600282049050600182168061033c57607f821691505b60208210810361034f5761034e6102f8565b5b50919050565b5f6040820190506103685f830185610202565b6103756020830184610202565b939250505056fea2646970667358221220f4cac414fc49fe0752ca80b0963a0c6ee7f6075b34293fc4879f950b9d6f8dd964736f6c634300081e0033"
)

func deployByEthClient() {
	fail := func(step string, err error) {
		log.Printf("FAILED at %s: %v", step, err)
		debug.PrintStack()
		log.Fatal(err)
	}
	// 连接到以太坊网络（这里使用 sepolia 测试网络作为示例）
	client, err := ethclient.Dial("https://sepolia.infura.io/v3/efe70fa30ffd4eb08d17e189753d80b9")
	if err != nil {
		fail("dial sepolia", err)
	}

	// 创建私钥（在实际应用中，您应该使用更安全的方式来管理私钥）
	privateKey, err := crypto.HexToECDSA("fc56c167e6b531c6752ee8a5458148753a08a1ec3944b46af7c578e073424e97")
	if err != nil {
		fail("parse private key", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fail("cast public key", fmt.Errorf("error casting public key to ECDSA"))
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		fail("get nonce", err)
	}

	// 解码创建字节码（注意：必须是 creation bytecode，而非 runtime bytecode）
	creationBytecode, err := hex.DecodeString(contractBytecode)
	if err != nil {
		fail("decode creation bytecode", err)
	}

	// 编码构造参数（需与合约构造函数一致）
	input := "1.0"
	// 从 abigen 元数据读取 ABI，失败则退化到常量 ABI
	var parsedAbi *abi.ABI
	if a, e := store.StoreMetaData.GetAbi(); e == nil {
		parsedAbi = a
	} else {
		abiReader := strings.NewReader(store.StoreABI)
		av, e2 := abi.JSON(abiReader)
		if e2 != nil {
			fail("parse ABI from const", e2)
		}
		parsedAbi = &av
	}

	constructorArgs, err := parsedAbi.Constructor.Inputs.Pack(input)
	if err != nil {
		fail("pack constructor args", err)
	}
	// 创建字节码 = 创建字节码 + 构造参数编码
	initCode := append(creationBytecode, constructorArgs...)

	// 使用 EIP-1559 费用模型并估算 gas
	ctx := context.Background()
	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		fail("suggest tip cap", err)
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		fail("get latest header", err)
	}
	base := header.BaseFee
	if base == nil {
		base = big.NewInt(0)
	}
	feeCap := new(big.Int).Add(tipCap, new(big.Int).Mul(base, big.NewInt(2)))

	// 估算部署所需的 gasLimit
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddress,
		To:    nil,
		Value: big.NewInt(0),
		Data:  initCode,
	})
	if err != nil {
		fail("estimate gas", err)
	}

	// 获取链ID（用于签名和构造交易）
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fail("get chain id", err)
	}

	// 构造 EIP-1559 部署交易
	dynTx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		Value:     big.NewInt(0),
		Data:      initCode,
	}
	tx := types.NewTx(dynTx)

	// 签名交易：EIP-1559 动态费用交易需使用 London 或 Latest 签名器
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		fail("sign tx", err)
	}

	// 发送交易（若节点不支持 EIP-1559，则回退到 Legacy 交易）
	ctxSend := context.Background()
	finalTx := signedTx
	if err = client.SendTransaction(ctxSend, signedTx); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "type not supported") || strings.Contains(msg, "unsupported transaction type") {
			// 回退为 Legacy 交易
			gasPrice, gerr := client.SuggestGasPrice(ctxSend)
			if gerr != nil {
				fail("fallback suggest gas price", gerr)
			}
			// 重新估算 legacy 模式的 gas
			legacyGas, lgerr := client.EstimateGas(ctxSend, ethereum.CallMsg{From: fromAddress, To: nil, Value: big.NewInt(0), Data: initCode, GasPrice: gasPrice})
			if lgerr != nil {
				fail("fallback estimate gas", lgerr)
			}
			legacy := types.NewContractCreation(nonce, big.NewInt(0), legacyGas, gasPrice, initCode)
			legacySigned, serr := types.SignTx(legacy, types.NewEIP155Signer(chainID), privateKey)
			if serr != nil {
				fail("fallback sign", serr)
			}
			if serr = client.SendTransaction(ctxSend, legacySigned); serr != nil {
				fail("send legacy tx", serr)
			}
			finalTx = legacySigned
			fmt.Printf("Transaction sent (legacy fallback): %s (type=%d, gas=%d)\n", finalTx.Hash().Hex(), finalTx.Type(), finalTx.Gas())
		} else {
			fail("send tx", err)
		}
	} else {
		fmt.Printf("Transaction sent: %s (type=%d, gas=%d)\n", finalTx.Hash().Hex(), finalTx.Type(), finalTx.Gas())
	}

	// 等待交易被挖矿
	receipt, err := waitForReceipt(client, finalTx.Hash())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Contract deployed at: %s\n", receipt.ContractAddress.Hex())
}

func waitForReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := client.TransactionReceipt(context.Background(), txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		// 等待一段时间后再次查询
		time.Sleep(1 * time.Second)
	}
}
