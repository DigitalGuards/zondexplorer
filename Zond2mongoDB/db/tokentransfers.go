package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"Zond2mongoDB/configs"
	"Zond2mongoDB/models"
	"Zond2mongoDB/rpc"
)

func normalizeAddress(addr string) string {
	
	if addr == "" {
		return "0x0"
	}
	return addr

}

func StoreTokenTransfer(transfer models.TokenTransfer) error {

	collection := configs.GetTokenTransfersCollection()
	ctx := context.Background()

	transfer.From = normalizeAddress(transfer.From)
	transfer.To = normalizeAddress(transfer.To)

	configs.Logger.Info("Storing token transfer",
		zap.String("token", transfer.TokenSymbol),
		zap.String("from", transfer.From),
		zap.String("to", transfer.To),
		zap.String("txHash", transfer.TxHash))

	_, err := collection.InsertOne(ctx, transfer)
	if err != nil {
		configs.Logger.Error("Failed to store token transfer",
			zap.String("txHash", transfer.TxHash),
			zap.Error(err))
		return err
	}

	configs.Logger.Info("Token transfer stored",
		zap.String("txHash", transfer.TxHash),
		zap.String("token", transfer.TokenSymbol))
	return nil

}

func getTransfers(filter bson.M, skip, limit int64) ([]models.TokenTransfer, error) {

	collection := configs.GetCollection(configs.DB, "tokenTransfers")
	ctx := context.Background()

	opts := options.Find().SetSort(bson.D{{Key: "blockNumber", Value: -1}}).SetSkip(skip).SetLimit(limit)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transfers []models.TokenTransfer
	err = cursor.All(ctx, &transfers)
	return transfers, err

}

func GetTokenTransfersByContract(contractAddress string, skip, limit int64) ([]models.TokenTransfer, error) {

	return getTransfers(bson.M{"contractAddress": contractAddress}, skip, limit)

}

func GetTokenTransfersByAddress(address string, skip, limit int64) ([]models.TokenTransfer, error) {

	return getTransfers(bson.M{"$or": []bson.M{{"from": address}, {"to": address}}}, skip, limit)

}

func TokenTransferExists(txHash, contractAddress, from, to string) (bool, error) {

	collection := configs.GetCollection(configs.DB, "tokenTransfers")
	ctx := context.Background()

	filter := bson.M{
		"txHash":          txHash,
		"contractAddress": contractAddress,
		"from":            from,
		"to":              to,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		configs.Logger.Error("Failed to check if token transfer exists", zap.String("txHash", txHash), zap.Error(err))
		return false, err
	}
	return count > 0, nil

}

func InitializeTokenTransfersCollection() error {

	collection := configs.GetTokenTransfersCollection()
	ctx := context.Background()

	configs.Logger.Info("Initializing tokenTransfers collection and indexes")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "contractAddress", Value: 1}, {Key: "blockNumber", Value: 1}}, Options: options.Index().SetName("contract_block_idx")} ,
		{Keys: bson.D{{Key: "from", Value: 1}, {Key: "blockNumber", Value: 1}}, Options: options.Index().SetName("from_block_idx")} ,
		{Keys: bson.D{{Key: "to", Value: 1}, {Key: "blockNumber", Value: 1}}, Options: options.Index().SetName("to_block_idx")} ,
		{Keys: bson.D{{Key: "txHash", Value: 1}}, Options: options.Index().SetName("txHash_idx").SetUnique(true)} ,
	}

	_, err := collection.Indexes().DropAll(ctx)
	if err != nil {
		configs.Logger.Warn("Failed to drop existing indexes", zap.Error(err))
	}
	_, err = collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		configs.Logger.Error("Failed to create indexes", zap.Error(err))
		return err
	}

	configs.Logger.Info("TokenTransfers collection initialized")
	return nil

}

func InitializeTokenBalancesCollection() error {

	collection := configs.GetTokenBalancesCollection()
	ctx := context.Background()

	configs.Logger.Info("Initializing tokenBalances collection and indexes")

	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "contractAddress", Value: 1}, {Key: "address", Value: 1}}, Options: options.Index().SetName("contract_address_idx").SetUnique(true)} ,
		{Keys: bson.D{{Key: "address", Value: 1}}, Options: options.Index().SetName("address_idx")} ,
		{Keys: bson.D{{Key: "contractAddress", Value: 1}}, Options: options.Index().SetName("contract_idx")} ,
	}

	_, err := collection.Indexes().DropAll(ctx)
	if err != nil {
		configs.Logger.Warn("Failed to drop indexes", zap.Error(err))
	}
	_, err = collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		configs.Logger.Error("Failed to create indexes", zap.Error(err))
		return err
	}

	configs.Logger.Info("TokenBalances collection initialized")
	return nil

}

func ProcessBlockTokenTransfers(blockNumber string, blockTimestamp string) error {
	transferEventSignature := "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"

	configs.Logger.Info("Searching for token transfers", zap.String("blockNumber", blockNumber), zap.String("eventSignature", transferEventSignature))
	response, err := rpc.ZondGetBlockLogs(blockNumber, []string{transferEventSignature})
	if err != nil {
		configs.Logger.Error("Failed to get logs for block", zap.String("blockNumber", blockNumber), zap.Error(err))
		return err
	}

	if response == nil || len(response.Result) == 0 {
		configs.Logger.Debug("No token transfer logs found in block", zap.String("blockNumber", blockNumber))
		return nil
	}

	configs.Logger.Info("Found potential token transfer logs", zap.String("blockNumber", blockNumber), zap.Int("logCount", len(response.Result)))
	tokenTransfersFound := 0

	for _, log := range response.Result {
		if len(log.Topics) < 3 {
			configs.Logger.Debug("Skipping log with insufficient topics", zap.String("txHash", log.TransactionHash), zap.Int("topicCount", len(log.Topics)))
			continue
		}

		contractAddress := log.Address
		configs.Logger.Debug("Processing potential token transfer", zap.String("contractAddress", contractAddress), zap.String("txHash", log.TransactionHash))

		name, symbol, decimals, isTokenRPC := rpc.GetTokenInfo(contractAddress)
		if isTokenRPC {
			txDetails, _ := rpc.GetTxDetailsByHash(log.TransactionHash)
			existingContract, _ := GetContract(contractAddress)

			contractUpdate := models.ContractInfo{
				Address:   contractAddress,
				Status:    "0x1",
				IsToken:   true,
				Name:      name,
				Symbol:    symbol,
				Decimals:  decimals,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
				CreationBlockNumber: blockNumber,
				CreationTransaction: log.TransactionHash,
				CreatorAddress:      txDetails.From,
			}

			if existingContract != nil {
				if existingContract.CreatorAddress != "" {
					contractUpdate.CreatorAddress = existingContract.CreatorAddress
				}
				if existingContract.CreationTransaction != "" {
					contractUpdate.CreationTransaction = existingContract.CreationTransaction
				}
				if existingContract.CreationBlockNumber != "" {
					contractUpdate.CreationBlockNumber = existingContract.CreationBlockNumber
				}
				if existingContract.ContractCode != "" {
					contractUpdate.ContractCode = existingContract.ContractCode
				}
			}

			totalSupply, supplyErr := rpc.GetTokenTotalSupply(contractAddress)
			if supplyErr == nil {
				contractUpdate.TotalSupply = totalSupply
			}

			_ = StoreContract(contractUpdate)
		}

		contract, err := GetContract(contractAddress)
		if err != nil || !contract.IsToken {
			configs.Logger.Debug("Skipping non-token or failed to get from DB", zap.String("address", contractAddress), zap.Error(err))
			continue
		}

		from := "0x" + rpc.TrimLeftZeros(log.Topics[1][26:])
		to := "0x" + rpc.TrimLeftZeros(log.Topics[2][26:])
		amount := log.Data

		exists, err := TokenTransferExists(log.TransactionHash, contractAddress, from, to)
		if err != nil || exists {
			configs.Logger.Debug("Skipping existing transfer or error", zap.String("txHash", log.TransactionHash), zap.Error(err))
			continue
		}

		if from == "" {
			from = "0x0"
		}
		if to == "" {
			to = "0x0"
		}

		transfer := models.TokenTransfer{
			ContractAddress: contractAddress,
			From:            from,
			To:              to,
			Amount:          amount,
			BlockNumber:     blockNumber,
			TxHash:          log.TransactionHash,
			Timestamp:       blockTimestamp,
			TokenSymbol:     contract.Symbol,
			TokenDecimals:   contract.Decimals,
			TokenName:       contract.Name,
			TransferType:    "event",
		}

		err = StoreTokenTransfer(transfer)
		if err != nil {
			configs.Logger.Error("Failed to store token transfer", zap.String("txHash", log.TransactionHash), zap.Error(err))
			continue
		}
		tokenTransfersFound++

		_ = StoreTokenBalance(contractAddress, from, amount, blockNumber)
		_ = StoreTokenBalance(contractAddress, to, amount, blockNumber)
	}

	configs.Logger.Info("Finished processing token transfers", zap.String("blockNumber", blockNumber), zap.Int("transfersProcessed", tokenTransfersFound))
	return nil
}

