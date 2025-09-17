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

func weiToEther(wei *big.Int) *big.Float {
	weiPerEth := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	weiFloat := new(big.Float).SetInt(wei)
	ethValue := new(big.Float).Quo(weiFloat, new(big.Float).SetInt(weiPerEth))
	return ethValue
}
