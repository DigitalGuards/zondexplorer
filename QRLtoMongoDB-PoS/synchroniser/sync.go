package synchroniser

import (
	"QRLtoMongoDB-PoS/configs"
	"QRLtoMongoDB-PoS/db"
	L "QRLtoMongoDB-PoS/logger"
	"QRLtoMongoDB-PoS/models"
	"QRLtoMongoDB-PoS/rpc"
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger = L.FileLogger(configs.Filename)

type Data struct {
	blockData    []interface{}
	blockNumbers []int
}

func Sync() {
	var bNum uint64 = db.GetLatestBlockNumberFromDB() + 1
	logger.Info("Starting sync from block number", zap.Uint64("block", bNum))

	wg := sync.WaitGroup{}

	max, err := rpc.GetLatestBlock()
	if err != nil {
		logger.Error("Failed to get latest block", zap.Error(err))
		return
	}
	logger.Info("Latest block from network", zap.Uint64("block", max))

	// Create a buffered channel of read only channels, with length 8.
	producers := make(chan (<-chan Data), 8)
	logger.Info("Initialized producer channels")

	// Start the consumer.
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer(os.Stdout, producers)
	}()
	logger.Info("Started consumer process")

	// Start producers in correct order.
	for i, size := int(bNum), 8; i < int(max); i += size {
		// Send the producer's channel over to the consumer.
		producers <- producer(i, i+size)
		logger.Info("Processing block range",
			zap.Int("from", i),
			zap.Int("to", i+size))
	}
	// No more producers will be started, so close the channel.
	close(producers)

	// Wait for consumer!
	wg.Wait()
	logger.Info("Initial sync completed successfully!")

	logger.Info("Calculating daily transaction volume...")
	db.GetDailyTransactionVolume()

	logger.Info("Starting continuous block monitoring...")
	singleBlockInsertion()
}

func consumer(w io.Writer, ch <-chan (<-chan Data)) {
	// Consume the producer channels.
	for Datas := range ch {
		// Consume the Datas.
		for i := range Datas {
			if i.blockData == nil || len(i.blockData) == 0 {
				continue
			}
			// Do stuff with the Datas, in order.
			db.InsertManyBlockDocuments(i.blockData)
			logger.Info("Inserted block batch",
				zap.Int("count", len(i.blockData)))

			for x := 0; x < len(i.blockNumbers); x++ {
				db.ProcessTransactions(i.blockData[x])
			}
			logger.Info("Processed transactions for blocks",
				zap.Ints("block_numbers", i.blockNumbers))
		}
	}
}

func producer(start int, end int) <-chan Data {
	// Create a channel which we will send our data.
	Datas := make(chan Data, 8)

	var blockData []interface{}
	var blockNumbers []int

	// Start the goroutine that produces data.
	go func(ch chan<- Data) {
		defer close(ch)

		// Produce data.
		for i := start; i < end; i++ {
			data := rpc.GetBlockByNumberMainnet(uint64(i))
			if data == "" {
				logger.Warn("Empty block data received", zap.Int("block", i))
				continue
			}

			var Zond models.Zond
			err := json.Unmarshal([]byte(data), &Zond)
			if err != nil {
				logger.Error("Failed to unmarshal block data",
					zap.Int("block", i),
					zap.Error(err))
				continue
			}

			if Zond.PreResult.ParentHash != "" {
				ZondNew := db.ConvertModelsUint64(Zond)
				blockData = append(blockData, ZondNew)
				blockNumbers = append(blockNumbers, i)
			}
		}
		if len(blockData) > 0 {
			ch <- Data{blockData: blockData, blockNumbers: blockNumbers}
		}
	}(Datas)

	// Return back to caller.
	return Datas
}

func processInitialBlock() {
	logger.Info("Processing genesis block (0)")
	blockData := rpc.GetBlockByNumberMainnet(0)
	if blockData == "" {
		logger.Error("Failed to get genesis block data")
		return
	}

	var Zond models.Zond
	err := json.Unmarshal([]byte(blockData), &Zond)
	if err != nil {
		logger.Error("Failed to unmarshal genesis block", zap.Error(err))
		return
	}
	ZondNew := db.ConvertModelsUint64(Zond)
	db.InsertBlockDocument(ZondNew)
	db.ProcessTransactions(ZondNew)
	logger.Info("Genesis block processed successfully")
}

func processSubsequentBlocks(sum uint64, latestBlockNumber uint64) uint64 {
	blockData := rpc.GetBlockByNumberMainnet(sum)
	if blockData == "" {
		logger.Warn("Empty block data received", zap.Uint64("block", sum))
		return sum
	}

	var Zond models.Zond
	err := json.Unmarshal([]byte(blockData), &Zond)
	if err != nil {
		logger.Error("Failed to unmarshal block",
			zap.Uint64("block", sum),
			zap.Error(err))
		return sum
	}

	if Zond.PreResult.ParentHash != "" {
		ZondNew := db.ConvertModelsUint64(Zond)
		previousHash := ZondNew.Result.ParentHash
		if previousHash == db.GetLatestBlockHashHeaderFromDB(db.GetLatestBlockNumberFromDB()) {
			processBlockAndUpdateValidators(sum, ZondNew, previousHash)
			logger.Info("Block processed successfully", zap.Uint64("block", sum))
			return sum + 1
		} else {
			// Chain reorganization detected
			logger.Warn("Chain reorganization detected, rolling back and finding fork point",
				zap.Uint64("block", sum))

			// Walk back through the chain until we find a matching block
			var lastValidBlock uint64
			for blockNum := sum - 1; blockNum > 0; blockNum-- {
				prevBlockData := rpc.GetBlockByNumberMainnet(blockNum)
				if prevBlockData == "" {
					continue
				}

				var prevZond models.Zond
				if err := json.Unmarshal([]byte(prevBlockData), &prevZond); err != nil {
					continue
				}

				prevZondNew := db.ConvertModelsUint64(prevZond)
				dbHash := db.GetLatestBlockHashHeaderFromDB(blockNum)

				if dbHash == prevZondNew.Result.Hash {
					lastValidBlock = blockNum
					break
				}

				logger.Info("Rolling back block", zap.Uint64("block", blockNum))
				db.Rollback(blockNum)
			}

			logger.Info("Found last valid block, resuming sync",
				zap.Uint64("last_valid_block", lastValidBlock))

			// Return next block to process
			return lastValidBlock + 1
		}
	}
	return sum
}

func processBlockAndUpdateValidators(sum uint64, ZondNew models.ZondDatabaseBlock, previousHash string) {
	db.InsertBlockDocument(ZondNew)
	db.ProcessTransactions(ZondNew)
	db.UpdateValidators(sum, previousHash)
}

func runPeriodicTask(task func(), interval time.Duration, taskName string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Recovered from panic in periodic task",
					zap.String("task", taskName),
					zap.Any("error", r))
			}
		}()

		logger.Info("Starting periodic task",
			zap.String("task", taskName),
			zap.Duration("interval", interval))

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Recovered from panic in periodic task tick",
							zap.String("task", taskName),
							zap.Any("error", r))
					}
				}()
				logger.Info("Running periodic task", zap.String("task", taskName))
				task()
				logger.Info("Completed periodic task", zap.String("task", taskName))
			}()
		}
	}()
}

func singleBlockInsertion() {
	sum := db.GetLatestBlockNumberFromDB() + 1
	logger.Info("Starting continuous block monitoring", zap.Uint64("from_block", sum))

	latestBlockNumber, err := rpc.GetLatestBlock()
	if err != nil {
		logger.Error("Failed to get latest block", zap.Error(err))
		return
	}
	logger.Info("Latest network block", zap.Uint64("block", latestBlockNumber))

	isCollectionsExist := db.IsCollectionsExist()

	// Block processing goroutine
	runPeriodicTask(func() {
		if !isCollectionsExist {
			processInitialBlock()
			isCollectionsExist = true
		} else {
			if sum <= latestBlockNumber {
				sum = processSubsequentBlocks(sum, latestBlockNumber)
			} else {
				latestBlockNumber, err = rpc.GetLatestBlock()
				if err != nil {
					logger.Error("Failed to get latest block", zap.Error(err))
				}
			}
		}
	}, 60*time.Second, "block_processing")

	// Data update goroutine
	runPeriodicTask(func() {
		func() {
			logger.Info("Updating CoinGecko data...")
			db.PeriodicallyUpdateCoinGeckoData()
		}()
		func() {
			logger.Info("Counting wallets...")
			db.CountWallets()
		}()
		func() {
			logger.Info("Calculating daily transaction volume...")
			db.GetDailyTransactionVolume()
		}()
	}, 5*time.Minute, "data_updates")

	select {}
}
