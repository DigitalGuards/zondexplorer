package routes

import (
	"backendAPI/db"
	"backendAPI/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoute(router *gin.Engine) {
	// Add pending transactions endpoint with pagination
	router.GET("/pending-transactions", func(c *gin.Context) {
		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		result, err := db.GetPendingTransactions(page, limit)
		if err != nil {
			log.Printf("Error fetching pending transactions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch pending transactions: %v", err),
			})
			return
		}

		// Log the result for debugging
		log.Printf("Found %d pending transactions", len(result.Transactions))

		// Return empty array instead of null for transactions
		if result.Transactions == nil {
			result.Transactions = make([]models.PendingTransaction, 0)
		}

		c.JSON(http.StatusOK, result)
	})

	// Add endpoint for fetching a specific pending transaction
	router.GET("/pending-transaction/:hash", func(c *gin.Context) {
		hash := c.Param("hash")
		transaction, err := db.GetPendingTransactionByHash(hash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If not found in pending, check if it's in the transactions collection
		if transaction == nil {
			// Check if transaction exists in the transactions collection
			tx, err := db.GetTransactionByHash(hash)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			if tx != nil {
				// Transaction is mined - return formatted response
				c.JSON(http.StatusOK, gin.H{
					"transaction": gin.H{
						"hash":        tx.Hash,        // Already has 0x prefix
						"status":      "mined",
						"blockNumber": tx.BlockNumber,
						"timestamp":   time.Now().Unix(),
					},
				})
				return
			}

			// Transaction not found in either collection
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Transaction not found",
				"details": "This transaction is no longer in the mempool. It may have been dropped or replaced.",
			})
			return
		}

		// If transaction is mined, delete it from pending collection
		if transaction.Status == "mined" {
			if err := db.DeleteMinedTransaction(hash); err != nil {
				// Log error but don't fail the request
				log.Printf("Error deleting mined transaction %s: %v\n", hash, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{"transaction": transaction})
	})

	router.GET("/overview", func(c *gin.Context) {
		// Get market cap with default value
		marketCap := db.GetMarketCap()

		// Get current price with default value
		currentPrice := db.GetCurrentPrice()

		// Get wallet count with default value
		walletCount := db.GetWalletCount()

		// Get circulating supply with default value
		circulating := db.ReturnTotalCirculatingSupply()
		if circulating == "" {
			circulating = "0" // Default value when no data is available
		}

		// Get daily transaction volume with default value
		volume := db.ReturnDailyTransactionsVolume()

		// Get validator count
		validatorCount, err := db.CountValidators()
		if err != nil {
			validatorCount = 0
		}

		// Get contract count
		contractCount, err := db.CountContracts()
		if err != nil {
			contractCount = 0
		}

		// Return response with default values if data isn't available
		c.JSON(http.StatusOK, gin.H{
			"marketcap":      marketCap,      // Returns 0 if not available
			"currentPrice":   currentPrice,   // Returns 0 if not available
			"countwallets":   walletCount,    // Returns 0 if not available
			"circulating":    circulating,    // Returns "0" if not available
			"volume":         volume,         // Returns 0 if not available
			"validatorCount": validatorCount, // Returns 0 if not available
			"contractCount":  contractCount,  // Returns 0 if not available
			"status": gin.H{
				"syncing":         true, // Indicate that data is still being synced
				"dataInitialized": marketCap > 0 || currentPrice > 0 || walletCount > 0 || circulating != "0" || volume > 0,
			},
		})
	})

	router.POST("/getBalance", func(c *gin.Context) {
		address := c.PostForm("address")

		balance, message := db.GetBalance(address)
		if message == "" {
			c.JSON(http.StatusOK, gin.H{
				"balance": balance,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"balance": message,
			})
		}
	})

	router.GET("/txs", func(c *gin.Context) {
		pageStr := c.Query("page")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			fmt.Println(err)
		}

		txs, err := db.ReturnTransactionsNetwork(page)
		if err != nil {
			fmt.Println(err)
		}

		// Transaction count for the address
		countTransactions, err := db.CountTransactionsNetwork()
		if err != nil {
			fmt.Println(err)
		}

		// Get latest block for confirmation count
		latestBlock, err := db.ReturnLatestBlock()
		if err != nil {
			fmt.Println(err)
		}

		var latestBlockNumber uint64
		if len(latestBlock) > 0 {
			latestBlockNumber = latestBlock[0].Result.Number
		}

		c.JSON(http.StatusOK, gin.H{
			"txs":         txs,
			"total":       countTransactions,
			"latestBlock": latestBlockNumber,
		})
	})

	router.GET("/walletdistribution/:query", func(c *gin.Context) {
		value := c.Param("query")
		wallets, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		query, err := db.ReturnWalletDistribution(wallets)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"response": query})
	})

	router.GET("/address/aggregate/:query", func(c *gin.Context) {
		param := c.Param("query")

		// Single Address data
		addressData, err := db.ReturnSingleAddress(param)
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error querying address: %v", err)})
			return
		}

		// Transaction count for the address
		countTransactions, err := db.CountTransactions(param)
		if err != nil {
			fmt.Printf("Error counting transactions: %v\n", err)
		}

		// Rank of the address
		rank, err := db.ReturnRankAddress(param)
		if err != nil {
			fmt.Printf("Error getting rank: %v\n", err)
		}

		// All transactions by the address
		TransactionsByAddress, err := db.ReturnAllTransactionsByAddress(param)
		if err != nil {
			fmt.Printf("Error getting transactions: %v\n", err)
		}

		// All internal transactions by the address
		InternalTransactionsByAddress, err := db.ReturnAllInternalTransactionsByAddress(param)
		if err != nil {
			fmt.Printf("Error getting internal transactions: %v\n", err)
		}

		// Contract code (if applicable)
		contractCodeData, err := db.ReturnContractCode(param)
		// Don't treat missing contract code as an error since not all addresses are contracts
		if err != nil && err != mongo.ErrNoDocuments {
			fmt.Printf("Error getting contract code: %v\n", err)
		}

		// Get latest block for confirmation count
		latestBlock, err := db.ReturnLatestBlock()
		if err != nil {
			fmt.Printf("Error getting latest block: %v\n", err)
		}

		var latestBlockNumber uint64
		if len(latestBlock) > 0 {
			latestBlockNumber = latestBlock[0].Result.Number
		}

		// Response aggregation
		c.JSON(http.StatusOK, gin.H{
			"address":                          addressData,
			"transactions_count":               countTransactions,
			"rank":                             rank,
			"transactions_by_address":          TransactionsByAddress,
			"internal_transactions_by_address": InternalTransactionsByAddress,
			"contract_code":                    contractCodeData,
			"latestBlock":                      latestBlockNumber,
		})
	})

	router.GET("/tx/:query", func(c *gin.Context) {
		value := c.Param("query")
		query, err := db.ReturnSingleTransfer(value)
		if err != nil {
			fmt.Println(err)
		}

		latestBlock, err := db.ReturnLatestBlock()
		if err != nil {
			fmt.Println(err)
		}

		var latestBlockNumber uint64
		if len(latestBlock) > 0 {
			latestBlockNumber = latestBlock[0].Result.Number
		}

		c.JSON(http.StatusOK, gin.H{
			"response":    query,
			"latestBlock": latestBlockNumber,
		})
	})

	router.GET("/latestblock", func(c *gin.Context) {
		latestBlock, err := db.ReturnLatestBlock()
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"response": latestBlock,
		})
	})

	router.GET("/coinbase/:query", func(c *gin.Context) {
		value := c.Param("query")
		query, err := db.ReturnSingleTransfer(value)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"response": query})
	})

	router.GET("/richlist", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"richlist": db.ReturnRichlist()})
	})

	router.GET("/blocks", func(c *gin.Context) {
		pageStr := c.Query("page")
		limitStr := c.Query("limit")

		page, err := strconv.Atoi(pageStr)
		if err != nil {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 5 // Default to 5 blocks per page
		}

		blocks, err := db.ReturnLatestBlocks(page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch blocks"})
			return
		}

		countBlocks, err := db.CountBlocksNetwork()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count blocks"})
			return
		}

		// Limit total pages to 300
		maxPages := int64(300)
		maxBlocks := maxPages * int64(limit)
		if countBlocks > maxBlocks {
			countBlocks = maxBlocks
		}

		c.JSON(http.StatusOK, gin.H{
			"blocks": blocks,
			"total":  countBlocks,
		})
	})

	router.GET("/blocksizes", func(c *gin.Context) {
		query, err := db.ReturnBlockSizes()
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"response": query})
	})

	router.GET("/validators", func(c *gin.Context) {
		pageToken := c.Query("page_token")
		rawValidators, err := db.ReturnValidators(pageToken)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch validators: %v", err),
			})
			return
		}

		// Get the current epoch from the latest block
		latestBlock, err := db.ReturnLatestBlock()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get latest block"})
			return
		}
		currentEpoch := int(latestBlock[0].Result.Number / 128) // Each epoch is 128 blocks

		// Initialize response
		response := models.ValidatorResponse{
			Validators:    make([]models.Validator, 0),
			TotalStaked:   "0",
			NextPageToken: pageToken, // Pass through the page token
		}

		totalStaked := float64(0)

		processedValidators := make(map[string]bool)

		// Process validators by slot
		for _, slotValidators := range rawValidators.Resultvalidator.Validatorsbyslotnumber {
			// Process leader
			validatorEntry := models.Validator{
				Address:      slotValidators.Leader,
				Uptime:       100.0, // TODO: Calculate actual uptime from historical data
				Age:          currentEpoch,
				StakedAmount: "40000000000000000000000", // 40000 Quanta in Wei (18 decimal places)
				IsActive:     true,
			}
			response.Validators = append(response.Validators, validatorEntry)
			totalStaked += 40000000000000000000000

			// Process attestors
			for _, attestor := range slotValidators.Attestors {
				if _, exists := processedValidators[attestor]; exists {
					continue
				}
				processedValidators[attestor] = true

				validatorEntry := models.Validator{
					Address:      attestor,
					Uptime:       100.0,
					Age:          currentEpoch,
					StakedAmount: "40000000000000000000000", // 40000 Quanta in Wei (18 decimal places)
					IsActive:     true,
				}
				response.Validators = append(response.Validators, validatorEntry)
				totalStaked += 40000000000000000000000
			}
		}

		response.TotalStaked = fmt.Sprintf("%.0f", totalStaked)

		c.JSON(http.StatusOK, response)
	})

	router.GET("/transactions", func(c *gin.Context) {
		query, err := db.ReturnLatestTransactions()
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"response": query})
	})

	router.GET("/contracts", func(c *gin.Context) {
		// Parse pagination parameters
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "0"), 10, 64)
		limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "10"), 10, 64)
		search := c.Query("search")

		query, total, err := db.ReturnContracts(page, limit, search)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch contracts: %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"response": query,
			"total":    total,
		})
	})

	router.GET("/block/:query", func(c *gin.Context) {
		value := c.Param("query")
		intValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			fmt.Println(err)
		}
		query, err := db.ReturnSingleBlock(intValue)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(http.StatusOK, gin.H{"response": query})
	})
}
