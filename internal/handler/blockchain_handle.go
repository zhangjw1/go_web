package handler

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

type BlockChainHandler struct {
	config    *config.Config
	logger    *logger.Logger
	client    *ethclient.Client
	startTime time.Time
	// optional deps
	db        *database.Database
	cache     cache.CacheService
	messaging messaging.MessagingService
}

func NewBlockChainHandler(cfg *config.Config, log *logger.Logger, db *database.Database, cacheSvc cache.CacheService, msgSvc messaging.MessagingService) (*BlockChainHandler, error) {
	var client *ethclient.Client
	var err error

	// 只有在区块链功能启用时才初始化客户端
	if cfg.Blockchain.Enabled {
		client, err = ethclient.Dial(cfg.Blockchain.NetworkURL)
		if err != nil {
			log.Error("Failed to connect to blockchain network", "error", err, "url", cfg.Blockchain.NetworkURL)
			return nil, fmt.Errorf("failed to connect to blockchain network: %w", err)
		}
		log.Info("Successfully connected to blockchain network", "network", cfg.Blockchain.NetworkName, "url", cfg.Blockchain.NetworkURL)
	}

	return &BlockChainHandler{
		config:    cfg,
		logger:    log,
		client:    client,
		startTime: time.Now(),
		db:        db,
		cache:     cacheSvc,
		messaging: msgSvc,
	}, nil
}

// GetLatestBlockNumber 获取最新区块号
// @Summary 获取最新区块号
// @Description 获取以太坊网络的最新区块号
// @Tags blockchain
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回最新区块号与网络信息"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "获取区块号失败"
// @Router /api/v1/blockchain/block/latest [get]
func (h *BlockChainHandler) GetLatestBlockNumber(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.Blockchain.Timeout)*time.Second)
	defer cancel()

	blockNumber, err := h.client.BlockNumber(ctx)
	if err != nil {
		h.logger.Error("Failed to get latest block number", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get latest block number",
		})
		return
	}

	h.logger.Info("Retrieved latest block number", "blockNumber", blockNumber)
	c.JSON(http.StatusOK, gin.H{
		"blockNumber": blockNumber,
		"network":     h.config.Blockchain.NetworkName,
	})
}

// GetBlockByNumber 根据区块号获取区块信息
// @Summary 获取区块信息
// @Description 通过区块号获取指定区块的基础信息
// @Tags blockchain
// @Produce json
// @Param number path integer true "区块号"
// @Success 200 {object} map[string]interface{} "成功返回区块信息"
// @Failure 400 {object} map[string]interface{} "区块号格式错误"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "获取区块信息失败"
// @Router /api/v1/blockchain/block/{number} [get]
func (h *BlockChainHandler) GetBlockByNumber(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	blockNumberStr := c.Param("number")
	blockNumber, err := strconv.ParseUint(blockNumberStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid block number",
		})
		return
	}

	//	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.Blockchain.Timeout)*time.Second)
	//	defer cancel()

	block, err := h.client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		h.logger.Error("Failed to get block by number", "error", err, "blockNumber", blockNumber)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get block information",
		})
		return
	}

	for _, tx := range block.Transactions() {
		h.logger.Info("Retrieved transaction information", "txHash", tx.Hash().Hex())
	}

	blockInfo := gin.H{
		"number":       block.Number().Uint64(),
		"hash":         block.Hash().Hex(),
		"parentHash":   block.ParentHash().Hex(),
		"timestamp":    block.Time(),
		"gasLimit":     block.GasLimit(),
		"gasUsed":      block.GasUsed(),
		"difficulty":   block.Difficulty().String(),
		"size":         block.Size(),
		"transactions": len(block.Transactions()),
		"network":      h.config.Blockchain.NetworkName,
	}

	h.logger.Info("Retrieved block information", "blockNumber", blockNumber)
	c.JSON(http.StatusOK, blockInfo)
}

// GetTransactionByHash 根据交易哈希获取交易信息
// @Summary 获取交易信息
// @Description 通过交易哈希获取交易与收据信息
// @Tags blockchain
// @Produce json
// @Param hash path string true "交易哈希(0x开头的66位)"
// @Success 200 {object} map[string]interface{} "成功返回交易信息"
// @Failure 400 {object} map[string]interface{} "交易哈希格式错误"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "获取交易信息失败"
// @Router /api/v1/blockchain/transaction/{hash} [get]
func (h *BlockChainHandler) GetTransactionByHash(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	txHash := c.Param("hash")
	if !common.IsHexAddress(txHash) && len(txHash) != 66 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction hash",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.Blockchain.Timeout)*time.Second)
	defer cancel()

	hash := common.HexToHash(txHash)
	tx, isPending, err := h.client.TransactionByHash(ctx, hash)
	if err != nil {
		h.logger.Error("Failed to get transaction by hash", "error", err, "txHash", txHash)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get transaction information",
		})
		return
	}

	chainID, err := h.client.NetworkID(context.Background())

	var signer types.Signer
	switch tx.Type() {
	case types.LegacyTxType:
		signer = types.NewEIP155Signer(chainID)
	case types.AccessListTxType:
		signer = types.NewEIP2930Signer(chainID)
	case types.DynamicFeeTxType:
		signer = types.NewLondonSigner(chainID)
	default:
		fmt.Errorf("unsupported transaction type: %d", tx.Type())
	}
	fmt.Printf("tx.Type(): %v\n", tx.Type())

	sender, err := types.Sender(signer, tx)

	// 获取交易收据
	receipt, err := h.client.TransactionReceipt(ctx, hash)
	if err != nil {
		h.logger.Warn("Failed to get transaction receipt", "error", err, "txHash", txHash)
	}

	txInfo := gin.H{
		"sender":   sender.Hex(),
		"hash":     tx.Hash().Hex(),
		"nonce":    tx.Nonce(),
		"to":       tx.To().Hex(),
		"value":    tx.Value().String(),
		"gasLimit": tx.Gas(),
		"gasPrice": tx.GasPrice().String(),
		"data":     fmt.Sprintf("0x%x", tx.Data()),
		"pending":  isPending,
		"network":  h.config.Blockchain.NetworkName,
	}

	if tx.To() != nil {
		txInfo["to"] = tx.To().Hex()
	}

	if receipt != nil {
		txInfo["blockNumber"] = receipt.BlockNumber.Uint64()
		txInfo["blockHash"] = receipt.BlockHash.Hex()
		txInfo["gasUsed"] = receipt.GasUsed
		txInfo["status"] = receipt.Status
	}

	h.logger.Info("Retrieved transaction information", "txHash", txHash)
	c.JSON(http.StatusOK, txInfo)
}

// GetBalance 获取地址余额
// @Summary 获取地址余额
// @Description 获取地址在当前网络的余额(ETH)
// @Tags blockchain
// @Produce json
// @Param address path string true "以太坊地址"
// @Success 200 {object} map[string]interface{} "成功返回余额(ETH)"
// @Failure 400 {object} map[string]interface{} "地址格式错误"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "获取余额失败"
// @Router /api/v1/blockchain/balance/{address} [get]
func (h *BlockChainHandler) GetBalance(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	address := c.Param("address")
	if !common.IsHexAddress(address) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid address format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.Blockchain.Timeout)*time.Second)
	defer cancel()

	addr := common.HexToAddress(address)
	balance, err := h.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		h.logger.Error("Failed to get balance", "error", err, "address", address)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get balance",
		})
		return
	}

	h.logger.Info("Retrieved balance", "address", address, "balance", balance.String())
	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"balance": weiToEther(balance),
		"network": h.config.Blockchain.NetworkName,
	})
}

// GenerateWallet 生成新钱包(私钥/公钥/地址)
// @Summary 生成钱包
// @Description 生成新的椭圆曲线密钥对以及对应地址
// @Tags blockchain
// @Produce json
// @Success 200 {object} map[string]interface{} "成功返回私钥、公钥与地址"
// @Router /api/v1/blockchain/wallet/create [get]
func (h *BlockChainHandler) GenerateWallet(c *gin.Context) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate private key",
		})
		return
	}

	//私钥字节
	privateKeyBytes := crypto.FromECDSA(privateKey)
	//转成十六进制字符串，并删除前缀0x(这就是用于签署交易的私钥)
	s := hexutil.Encode(privateKeyBytes)[2:]
	fmt.Printf("Private key: %s\n", s)

	//生成对应的公钥
	publicKey := privateKey.Public()
	//publicKeyECDSA 是一个 ECDSA（椭圆曲线数字签名算法）公钥的 Go 语言表示
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)

	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate public key",
		})
		return
	}

	/*
		当使用 crypto.FromECDSAPub() 将 ECDSA 公钥转换为字节时，它会返回一个 65 字节的数组，其中：
		第1个字节（索引0）是前缀 0x04，表示这是一个未压缩的公钥格式
		接下来的32字节是公钥的 X 坐标
		最后的32字节是公钥的 Y 坐标
	*/
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKetStr := hexutil.Encode(publicKeyBytes)[4:]
	fmt.Println("from pubKey:", publicKetStr) // 去掉'0x04'

	//通过公钥生成对应的 地址
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Printf("Address: %s\n", address)

	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes[1:])
	fmt.Println(hexutil.Encode(hash.Sum(nil)[12:]))

	c.JSON(http.StatusOK, gin.H{
		"privateKey": s,
		"publicKey":  publicKetStr,
		"address":    address,
	})
}

// TransferEther 转账ETH
// @Summary 转账 ETH
// @Description 使用私钥从对应地址向目标地址发起一笔ETH转账
// @Tags blockchain
// @Produce json
// @Param privateKey path string true "十六进制私钥(不含0x)"
// @Param toAddress path string true "接收方以太坊地址"
// @Success 200 {object} map[string]interface{} "成功返回交易哈希等信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "签名或发送交易失败"
// @Router /api/v1/blockchain/transfer/{privateKey}/{toAddress} [get]
func (h *BlockChainHandler) TransferEther(c *gin.Context) {
	privateKeyStr := c.Param("privateKey")
	if privateKeyStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	//加载私钥，生成对应的 *ecdsa.PrivateKey对象
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		h.logger.Error("Failed to convert private key", "error", err)
	}

	//生成公钥信息
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		h.logger.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	//公钥生成地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	//获取nonce
	nonce, err := h.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		h.logger.Error("Failed to get nonce", "error", err)
	}

	gwei := big.NewInt(1000000000000000)
	gasLimit := uint64(21000)
	gasPrice, err := h.client.SuggestGasPrice(context.Background())
	if err != nil {
		h.logger.Error("Failed to get gas price", "error", err)
	}

	params := c.Param("toAddress")
	toAddress := common.HexToAddress(params)

	var data []byte
	transaction := types.NewTransaction(nonce, toAddress, gwei, gasLimit, gasPrice, data)

	chainID, err := h.client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), privateKey)

	if err != nil {
		h.logger.Error("Failed to sign transaction", "error", err)
	}

	err = h.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		h.logger.Error("Failed to send transaction", "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"txHash":   signedTx.Hash().Hex(),
		"network":  h.config.Blockchain.NetworkName,
		"from":     fromAddress.Hex(),
		"to":       toAddress.Hex(),
		"value":    weiToEther(gwei),
		"gasLimit": gasLimit,
	})
}

// TransferToken 转账ERC20
// @Summary 转账 TOKEN
// @Description 使用私钥从对应地址向目标地址发起一笔ERC20转账
// @Tags blockchain
// @Produce json
// @Param privateKey path string true "十六进制私钥(不含0x)"
// @Param toAddress path string true "接收方以太坊地址"
// @Param tokenAddress path string true "ERC20代币合约地址"
// @Param amount path string true "转账数量"
// @Success 200 {object} map[string]interface{} "成功返回交易哈希等信息"
// @Failure 400 {object} map[string]interface{} "参数错误"
// @Failure 500 {object} map[string]interface{} "签名或发送交易失败"
// @Router /api/v1/blockchain/transfer-token/{privateKey}/{toAddress}/{tokenAddress}/{amount} [get]
func (h *BlockChainHandler) TransferToken(c *gin.Context) {

	//获取私钥参数
	privateKeyStr := c.Param("privateKey")
	if privateKeyStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing private key",
		})
		return
	}

	//获取接收地址参数
	toAddressStr := c.Param("toAddress")
	if toAddressStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing to address",
		})
		return
	}

	//获取代币地址参数
	tokenAddressStr := c.Param("tokenAddress")
	if tokenAddressStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing token address",
		})
		return
	}

	//获取token数量参数
	amountStr := c.Param("amount")
	if amountStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing amount",
		})
		return
	}

	// 解析金额（假设传入的是代币数量，需要转换为最小单位）
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid amount format",
		})
		return
	}

	// 转换为代币的最小单位（假设18位小数）
	decimals := big.NewInt(18)
	multiplier := new(big.Int).Exp(big.NewInt(10), decimals, nil)
	amountInWei := new(big.Int).Mul(amount, multiplier)

	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to convert private key",
		})
		return
	}
	//通过私钥解析对应的发送方地址
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	//获取nonce
	nonce, err := h.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get nonce",
		})
	}

	//ETH发送值
	value := big.NewInt(0)

	//获取gas价格
	gasPrice, err := h.client.SuggestGasPrice(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get gas price",
		})
	}

	tokenAddress := common.HexToAddress(tokenAddressStr)
	toAddress := common.HexToAddress(toAddressStr)

	//定义调用合约的方法
	transferFnSignature := []byte("transfer(address,uint256)")
	//计算签名函数的Keccak-256哈希，并取前4个字节作为函数选择器
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	//这个method ID用于在智能合约中识别要调用的具体函数
	methodID := hash.Sum(nil)[:4]
	fmt.Printf("methodID: %x\n", hexutil.Encode(methodID))
	//将接收方地址和金额转换为32字节的十六进制字符串(左边补充0)
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Printf("paddedAddress: %x\n", hexutil.Encode(paddedAddress))
	//将金额转换为32字节的十六进制字符串(左边补充0)
	paddedAmount := common.LeftPadBytes(amountInWei.Bytes(), 32)
	fmt.Printf("paddedAmount: %x\n", hexutil.Encode(paddedAmount))
	//data就是发送给ERC20合约的完整交易数据，格式符合以太坊ABI编码规范。这是调用智能合约函数的标准方式
	data := append(methodID, paddedAddress...)
	data = append(data, paddedAmount...)

	//通过合约的执行需求计算大致的gas费用
	gasLimit, err := h.client.EstimateGas(context.Background(), ethereum.CallMsg{
		To:   &toAddress,
		Data: data,
	})
	fmt.Printf("gasLimit: %d\n", gasLimit)

	//发起一笔交易
	transaction := types.NewTransaction(nonce, tokenAddress, value, gasLimit, gasPrice, data)
	chainID, err := h.client.NetworkID(context.Background())
	if err != nil {
		h.logger.Error("Failed to get network ID", "error", err)
	}

	//使用私钥对交易签名
	signedTx, err := types.SignTx(transaction, types.NewEIP155Signer(chainID), privateKey)

	if err != nil {
		h.logger.Error("Failed to sign transaction", "error", err)
	}

	//发送交易
	err = h.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		h.logger.Error("Failed to send transaction", "error", err)
	}
	c.JSON(http.StatusOK, gin.H{
		"txHash":          signedTx.Hash().Hex(),
		"network":         h.config.Blockchain.NetworkName,
		"from":            fromAddress.Hex(),
		"to":              toAddress.Hex(),
		"value":           weiToEther(value),
		"gasLimit":        gasLimit,
		"tokenAddress":    tokenAddress.Hex(),
		"amount":          amount.String(),
		"amountFormatted": formatTokenBalance(amountInWei),
		"status":          "success",
	})

}

// GetTokenBalance 获取ERC20代币余额
// @Summary 获取ERC20代币余额
// @Description 查询指定地址在指定ERC20代币合约中的余额
// @Tags blockchain
// @Produce json
// @Param address path string true "钱包地址"
// @Param tokenAddress path string true "ERC20代币合约地址"
// @Success 200 {object} map[string]interface{} "成功返回代币余额信息"
// @Failure 400 {object} map[string]interface{} "参数格式错误"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "查询余额失败"
// @Router /api/v1/blockchain/token-balance/{address}/{tokenAddress} [get]
func (h *BlockChainHandler) GetTokenBalance(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	// 获取钱包地址参数
	addressStr := c.Param("address")
	if addressStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing wallet address",
		})
		return
	}

	// 获取代币合约地址参数
	tokenAddressStr := c.Param("tokenAddress")
	if tokenAddressStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing token address",
		})
		return
	}

	// 验证地址格式
	if !common.IsHexAddress(addressStr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address format",
		})
		return
	}

	if !common.IsHexAddress(tokenAddressStr) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid token address format",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.config.Blockchain.Timeout)*time.Second)
	defer cancel()

	// 转换为地址类型
	walletAddress := common.HexToAddress(addressStr)
	tokenAddress := common.HexToAddress(tokenAddressStr)

	// 构建ERC20 balanceOf函数调用数据
	// balanceOf(address) 的函数选择器
	balanceOfSignature := []byte("balanceOf(address)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(balanceOfSignature)
	methodID := hash.Sum(nil)[:4]

	// 将钱包地址填充为32字节
	paddedAddress := common.LeftPadBytes(walletAddress.Bytes(), 32)

	// 组合调用数据
	data := append(methodID, paddedAddress...)

	// 调用合约
	result, err := h.client.CallContract(ctx, ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		h.logger.Error("Failed to call token contract", "error", err, "address", addressStr, "tokenAddress", tokenAddressStr)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query token balance",
		})
		return
	}

	// 解析余额结果（32字节大端序）
	balance := new(big.Int).SetBytes(result)

	h.logger.Info("Retrieved token balance", "address", addressStr, "tokenAddress", tokenAddressStr, "balance", balance.String())

	c.JSON(http.StatusOK, gin.H{
		"address":          addressStr,
		"tokenAddress":     tokenAddressStr,
		"balance":          balance.String(),
		"balanceFormatted": formatTokenBalance(balance),
		"network":          h.config.Blockchain.NetworkName,
	})
}

// SubscribeBlock 订阅新区块事件
// @Summary 订阅新区块事件
// @Description 实时订阅以太坊网络的新区块事件，并输出区块基本信息
// @Tags blockchain
// @Produce json
// @Success 200 {object} map[string]interface{} "成功订阅并返回区块信息"
// @Failure 503 {object} map[string]interface{} "区块链服务未启用"
// @Failure 500 {object} map[string]interface{} "订阅失败"
// @Router /api/v1/blockchain/subscribe/block [get]
func (h *BlockChainHandler) SubscribeBlock(c *gin.Context) {
	if !h.config.Blockchain.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Blockchain service is disabled",
		})
		return
	}

	headers := make(chan *types.Header)
	sub, err := h.client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		h.logger.Error("Failed to subscribe new head", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to subscribe new block events",
			"message": "The blockchain node may not support subscription or the connection is not available",
		})
		return
	}
	defer sub.Unsubscribe()

	// 设置SSE响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 发送连接成功的初始消息
	c.SSEvent("connected", gin.H{
		"message": "Successfully connected to block subscription",
		"network": h.config.Blockchain.NetworkName,
	})
	c.Writer.Flush()

	for {
		select {
		case err := <-sub.Err():
			h.logger.Error("Subscription error", "error", err)
			// 通过SSE发送错误信息而不是直接返回JSON
			c.SSEvent("error", gin.H{
				"error":   "Subscription error occurred",
				"message": err.Error(),
			})
			return
		case header := <-headers:
			block, err := h.client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				h.logger.Error("Failed to get block by hash", "error", err, "hash", header.Hash().Hex())
				// 即使获取区块失败，也继续监听
				c.SSEvent("error", gin.H{
					"error":   "Failed to get block by hash",
					"message": err.Error(),
					"hash":    header.Hash().Hex(),
				})
				c.Writer.Flush()
				continue
			}

			blockInfo := gin.H{
				"hash":         block.Hash().Hex(),
				"number":       block.Number().Uint64(),
				"timestamp":    block.Time(),
				"nonce":        block.Nonce(),
				"transactions": len(block.Transactions()),
			}

			h.logger.Info("New block received", "blockInfo", blockInfo)
			c.SSEvent("block", blockInfo)
			c.Writer.Flush()
		}
	}
}

func weiToEther(wei *big.Int) *big.Float {
	weiPerEth := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	weiFloat := new(big.Float).SetInt(wei)
	ethValue := new(big.Float).Quo(weiFloat, new(big.Float).SetInt(weiPerEth))
	return ethValue
}

// formatTokenBalance 格式化代币余额，假设代币精度为18位小数
func formatTokenBalance(balance *big.Int) string {
	// 大多数ERC20代币使用18位小数，与ETH相同
	decimals := big.NewInt(18)
	divisor := new(big.Int).Exp(big.NewInt(10), decimals, nil)

	// 转换为浮点数进行除法运算
	balanceFloat := new(big.Float).SetInt(balance)
	divisorFloat := new(big.Float).SetInt(divisor)
	result := new(big.Float).Quo(balanceFloat, divisorFloat)

	return result.Text('f', 18)
}
