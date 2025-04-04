package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"Zond2mongoDB/configs"
	"Zond2mongoDB/models"
	"Zond2mongoDB/rpc"
)

// isZeroAddress checks if the provided address is a known zero address variant

func isZeroAddress(addr string) bool {
	addr = strings.ToLower(addr)
	return addr == "0x0" ||
     	   addr == "0x0000000000000000000000000000000000000000" ||
		   addr == "z0" ||
		   addr == "z0000000000000000000000000000000000000000"
}

// StoreTokenBalance updates the token balance for a given address

func StoreTokenBalance(contractAddress, holderAddress, amount, blockNumber string) error {

	if isZeroAddress(holderAddress) {
		configs.Logger.Info("Skipping token balance update for zero address",
			zap.String("holderAddress", holderAddress))
		return nil
	}

	collection := configs.GetTokenBalancesCollection()
	if collection == nil {
		configs.Logger.Error("Failed to get token balances collection")
		return fmt.Errorf("token balances collection is nil")
	}

	balance, err := rpc.GetTokenBalance(contractAddress, holderAddress)
	if err != nil {
		configs.Logger.Error("Failed to get token balance from RPC",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.Error(err))
		configs.Logger.Info("Using default zero balance after RPC failure")
		balance = "0"
	}

	update := bson.M{
		"$set": bson.M{
			"contractAddress": contractAddress,
			"holderAddress":   holderAddress,
			"balance":         balance,
			"blockNumber":     blockNumber,
			"updatedAt":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	filter := bson.M{
		"contractAddress": contractAddress,
		"holderAddress":   holderAddress,
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		configs.Logger.Error("Failed to update token balance in database",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.Error(err))
		return fmt.Errorf("failed to update token balance: %v", err)
	}

	configs.Logger.Info("Token balance update completed",
		zap.String("contractAddress", contractAddress),
		zap.String("holderAddress", holderAddress),
		zap.Int64("matchedCount", result.MatchedCount),
		zap.Int64("modifiedCount", result.ModifiedCount),
		zap.Int64("upsertedCount", result.UpsertedCount))
	return nil

}

// GetTokenBalance retrieves the current token balance for a holder

func GetTokenBalance(contractAddress, holderAddress string) (*models.TokenBalance, error) {

	collection := configs.GetTokenBalancesCollection()
	if collection == nil {
		return nil, fmt.Errorf("token balances collection is nil")
	}

	filter := bson.M{
		"contractAddress": contractAddress,
		"holderAddress":   holderAddress,
	}

	var balance models.TokenBalance
	err := collection.FindOne(context.Background(), filter).Decode(&balance)
	if err != nil {
		return nil, err
	}

	return &balance, nil

}

// GetTokenHolders retrieves all holders of a specific token

func GetTokenHolders(contractAddress string) ([]models.TokenBalance, error) {

	collection := configs.GetTokenBalancesCollection()
	if collection == nil {
		return nil, fmt.Errorf("token balances collection is nil")
	}

	filter := bson.M{"contractAddress": contractAddress}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var balances []models.TokenBalance
	err = cursor.All(context.Background(), &balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
	
}
