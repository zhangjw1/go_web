package handler

import (
	"context"
	"fmt"
	"go-web-starter/internal/config"
	"go-web-starter/internal/infrastructure/cache"
	"go-web-starter/internal/infrastructure/database"
	"go-web-starter/internal/infrastructure/logger"
	"go-web-starter/internal/infrastructure/messaging"
	"math/big"
	"net/http"
	"strconv"
	"time"

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
// @Description 获取最新区块号
// @Tags Blockchain
// @Produce json
// @Success 200 {object}
// @Router /blockchain/latest
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
// @Description 获取区块信息
// @Tags Blockchain
// @Produce json
// @Param number path string true "Block number"
// @Success 200 {object}
// @Router /blockchain/block/{number}
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

	// 获取交易收据
	receipt, err := h.client.TransactionReceipt(ctx, hash)
	if err != nil {
		h.logger.Warn("Failed to get transaction receipt", "error", err, "txHash", txHash)
	}

	txInfo := gin.H{
		"hash":     tx.Hash().Hex(),
		"nonce":    tx.Nonce(),
		"to":       "",
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
		"balance": balance.String(),
		"network": h.config.Blockchain.NetworkName,
	})
}
