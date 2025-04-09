package db

import (
	"Zond2mongoDB/configs"
	"Zond2mongoDB/models"
	"Zond2mongoDB/rpc"
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// StoreTokenBalance updates the token balance for a given address
func StoreTokenBalance(contractAddress string, holderAddress string, amount string, blockNumber string) error {
	configs.Logger.Info("Attempting to store token balance",
		zap.String("contractAddress", contractAddress),
		zap.String("holderAddress", holderAddress),
		zap.String("transferAmount", amount),
		zap.String("blockNumber", blockNumber))

	// Special handling for zero address
	if holderAddress == "0x0" ||
		holderAddress == "0x0000000000000000000000000000000000000000" ||
		holderAddress == "Z0" ||
		holderAddress == "Z0000000000000000000000000000000000000000" ||
		strings.ToLower(holderAddress) == "0x0000000000000000000000000000000000000000" {
		configs.Logger.Info("Skipping token balance update for zero address",
			zap.String("holderAddress", holderAddress))
		return nil
	}

	collection := configs.GetTokenBalancesCollection()
	if collection == nil {
		configs.Logger.Error("Failed to get token balances collection")
		return fmt.Errorf("token balances collection is nil")
	}

	// Get current balance from RPC with more robust error handling
	configs.Logger.Debug("Calling RPC to get current token balance")
	balance, err := rpc.GetTokenBalance(contractAddress, holderAddress)
	if err != nil {
		// If RPC fails, try to get the existing balance from database and calculate new balance
		configs.Logger.Warn("Failed to get token balance from RPC, attempting to calculate from existing record",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.Error(err))

		// Get existing balance from DB
		var existingBalance models.TokenBalance
		filter := bson.M{
			"contractAddress": contractAddress,
			"holderAddress":   holderAddress,
		}

		err = collection.FindOne(context.Background(), filter).Decode(&existingBalance)
		if err == nil && existingBalance.Balance != "" {
			// Convert existing balance and amount to big.Int
			existingBig := new(big.Int)
			amountBig := new(big.Int)

			// Try to parse existing balance
			_, success1 := existingBig.SetString(existingBalance.Balance, 10)

			// Try to parse amount (strip 0x prefix if present)
			amountStr := amount
			isNegative := false
			if strings.HasPrefix(amountStr, "-") {
				isNegative = true
				amountStr = strings.TrimPrefix(amountStr, "-")
			}

			if strings.HasPrefix(amountStr, "0x") {
				amountStr = amountStr[2:]
				_, success2 := amountBig.SetString(amountStr, 16)
				if !success2 {
					configs.Logger.Error("Failed to parse transfer amount as hex",
						zap.String("amount", amount))
					amountBig.SetInt64(0)
				}
			} else {
				_, success2 := amountBig.SetString(amountStr, 10)
				if !success2 {
					configs.Logger.Error("Failed to parse transfer amount as decimal",
						zap.String("amount", amount))
					amountBig.SetInt64(0)
				}
			}

			// If negative (sender of a transfer), subtract amount
			if isNegative {
				amountBig = amountBig.Neg(amountBig)
			}

			if success1 {
				// Calculate new balance based on whether this is a send or receive
				newBalance := new(big.Int).Add(existingBig, amountBig)
				// Ensure balance doesn't go negative
				if newBalance.Sign() < 0 {
					configs.Logger.Warn("Calculated negative balance, setting to zero",
						zap.String("contractAddress", contractAddress),
						zap.String("holderAddress", holderAddress),
						zap.String("existingBalance", existingBalance.Balance),
						zap.String("amount", amount))
					newBalance.SetInt64(0)
				}
				balance = newBalance.String()

				configs.Logger.Info("Calculated balance from existing record",
					zap.String("contractAddress", contractAddress),
					zap.String("holderAddress", holderAddress),
					zap.String("existingBalance", existingBalance.Balance),
					zap.String("amount", amount),
					zap.String("calculatedBalance", balance))
			} else {
				configs.Logger.Error("Failed to parse existing balance",
					zap.String("balance", existingBalance.Balance))
				balance = "0"
			}
		} else {
			configs.Logger.Info("No existing balance found, using default zero balance",
				zap.String("contractAddress", contractAddress),
				zap.String("holderAddress", holderAddress))
			balance = "0"
		}
	} else {
		configs.Logger.Info("Retrieved current token balance from RPC",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.String("balance", balance))
	}

	// Create update document
	update := bson.M{
		"$set": bson.M{
			"contractAddress": contractAddress,
			"holderAddress":   holderAddress,
			"balance":         balance,
			"blockNumber":     blockNumber,
			"updatedAt":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Update options
	opts := options.Update().SetUpsert(true)

	// Filter to find existing document
	filter := bson.M{
		"contractAddress": contractAddress,
		"holderAddress":   holderAddress,
	}

	// Perform upsert
	configs.Logger.Debug("Performing upsert operation for token balance")
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
func GetTokenBalance(contractAddress string, holderAddress string) (*models.TokenBalance, error) {
	collection := configs.GetTokenBalancesCollection()
	var balance models.TokenBalance

	filter := bson.M{
		"contractAddress": contractAddress,
		"holderAddress":   holderAddress,
	}

	err := collection.FindOne(context.Background(), filter).Decode(&balance)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// GetTokenHolders retrieves all holders of a specific token
func GetTokenHolders(contractAddress string) ([]models.TokenBalance, error) {
	collection := configs.GetTokenBalancesCollection()
	var balances []models.TokenBalance

	filter := bson.M{"contractAddress": contractAddress}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &balances)
	if err != nil {
		return nil, err
	}

	return balances, nil
}

// UpdateTokenBalancesForTransfer handles token balance updates for both sender and receiver in a transfer
func UpdateTokenBalancesForTransfer(contractAddress string, senderAddress string, receiverAddress string, amount string, blockNumber string) error {
	configs.Logger.Info("Updating token balances for transfer",
		zap.String("contractAddress", contractAddress),
		zap.String("senderAddress", senderAddress),
		zap.String("receiverAddress", receiverAddress),
		zap.String("amount", amount),
		zap.String("blockNumber", blockNumber))

	// Skip if either address is the zero address
	zeroAddresses := []string{"0x0", "0x0000000000000000000000000000000000000000", "Z0", "Z0000000000000000000000000000000000000000"}

	// Process sender balance (deduct tokens)
	if !containsAddress(zeroAddresses, senderAddress) {
		if err := updateSingleAddressBalance(contractAddress, senderAddress, amount, blockNumber, false); err != nil {
			configs.Logger.Error("Failed to update sender balance",
				zap.String("senderAddress", senderAddress),
				zap.Error(err))
			// Continue to update receiver even if sender update fails
		}
	}

	// Process receiver balance (add tokens)
	if !containsAddress(zeroAddresses, receiverAddress) {
		if err := updateSingleAddressBalance(contractAddress, receiverAddress, amount, blockNumber, true); err != nil {
			configs.Logger.Error("Failed to update receiver balance",
				zap.String("receiverAddress", receiverAddress),
				zap.Error(err))
			return err
		}
	}

	return nil
}

// Helper function to check if an address is in a list
func containsAddress(addresses []string, address string) bool {
	lowerAddress := strings.ToLower(address)
	for _, a := range addresses {
		if strings.ToLower(a) == lowerAddress {
			return true
		}
	}
	return false
}

// Helper function to update a single address balance
func updateSingleAddressBalance(contractAddress string, holderAddress string, amount string, blockNumber string, isAddition bool) error {
	collection := configs.GetTokenBalancesCollection()
	if collection == nil {
		configs.Logger.Error("Failed to get token balances collection")
		return fmt.Errorf("token balances collection is nil")
	}

	// First try to get balance from RPC
	balance, err := rpc.GetTokenBalance(contractAddress, holderAddress)
	if err == nil {
		// If RPC succeeds, use the returned balance directly
		configs.Logger.Info("Using direct RPC balance",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.String("balance", balance))
	} else {
		// If RPC fails, calculate based on existing balance
		configs.Logger.Warn("RPC balance retrieval failed, calculating from DB",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.Error(err))

		// Get existing balance from DB
		var existingBalance models.TokenBalance
		filter := bson.M{
			"contractAddress": contractAddress,
			"holderAddress":   holderAddress,
		}

		err = collection.FindOne(context.Background(), filter).Decode(&existingBalance)

		// Parse existing balance
		existingBig := new(big.Int)
		existingBalanceStr := "0"
		if err == nil && existingBalance.Balance != "" {
			existingBalanceStr = existingBalance.Balance
			_, success := existingBig.SetString(existingBalanceStr, 10)
			if !success {
				configs.Logger.Error("Failed to parse existing balance",
					zap.String("balance", existingBalanceStr))
				existingBig.SetInt64(0)
			}
		} else {
			existingBig.SetInt64(0)
		}

		// Parse amount
		amountBig := new(big.Int)
		amountStr := amount
		if strings.HasPrefix(amountStr, "0x") {
			amountStr = amountStr[2:]
			_, success := amountBig.SetString(amountStr, 16)
			if !success {
				configs.Logger.Error("Failed to parse amount as hex",
					zap.String("amount", amount))
				amountBig.SetInt64(0)
			}
		} else {
			_, success := amountBig.SetString(amountStr, 10)
			if !success {
				configs.Logger.Error("Failed to parse amount as decimal",
					zap.String("amount", amount))
				amountBig.SetInt64(0)
			}
		}

		// Calculate new balance
		newBalance := new(big.Int)
		if isAddition {
			newBalance.Add(existingBig, amountBig)
		} else {
			newBalance.Sub(existingBig, amountBig)
			// Ensure balance doesn't go negative
			if newBalance.Sign() < 0 {
				configs.Logger.Warn("Calculated negative balance, setting to zero",
					zap.String("contractAddress", contractAddress),
					zap.String("holderAddress", holderAddress),
					zap.String("existingBalance", existingBalanceStr),
					zap.String("amount", amount))
				newBalance.SetInt64(0)
			}
		}

		balance = newBalance.String()
		configs.Logger.Info("Calculated new balance",
			zap.String("contractAddress", contractAddress),
			zap.String("holderAddress", holderAddress),
			zap.String("existingBalance", existingBalanceStr),
			zap.String("amount", amount),
			zap.Bool("isAddition", isAddition),
			zap.String("newBalance", balance))
	}

	// Create update document
	update := bson.M{
		"$set": bson.M{
			"contractAddress": contractAddress,
			"holderAddress":   holderAddress,
			"balance":         balance,
			"blockNumber":     blockNumber,
			"updatedAt":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Update options
	opts := options.Update().SetUpsert(true)

	// Filter to find existing document
	filter := bson.M{
		"contractAddress": contractAddress,
		"holderAddress":   holderAddress,
	}

	// Perform upsert
	result, err := collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		configs.Logger.Error("Failed to update token balance",
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
