package configs

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func ConnectDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(EnvMongoURI()))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB")

	// Initialize collections with validators
	db := client.Database("qrldata-z")

	// Daily Transactions Volume
	volumeValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"volume", "timestamp"},
			"properties": bson.M{
				"volume": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
				"timestamp": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
			},
		},
	}
	ensureCollection(db, "dailyTransactionsVolume", volumeValidator)

	// CoinGecko Data
	coingeckoValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"marketCapUSD", "priceUSD", "lastUpdated"},
			"properties": bson.M{
				"marketCapUSD": bson.M{
					"bsonType":    "double",
					"description": "must be a double and is required",
				},
				"priceUSD": bson.M{
					"bsonType":    "double",
					"description": "must be a double and is required",
				},
				"lastUpdated": bson.M{
					"bsonType":    "date",
					"description": "must be a date and is required",
				},
			},
		},
	}
	ensureCollection(db, "coingecko", coingeckoValidator)

	// Wallet Count
	walletValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"count"},
			"properties": bson.M{
				"count": bson.M{
					"bsonType":    "long",
					"description": "must be a long/int64 and is required",
				},
			},
		},
	}
	ensureCollection(db, "walletCount", walletValidator)

	// Total Circulating Supply
	circulatingValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"circulating"},
			"properties": bson.M{
				"circulating": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
			},
		},
	}
	ensureCollection(db, "totalCirculatingSupply", circulatingValidator)

	// Token Balances
	tokenBalanceValidator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"contractAddress", "holderAddress", "balance", "blockNumber", "updatedAt"},
			"properties": bson.M{
				"contractAddress": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
				"holderAddress": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
				"balance": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
				"blockNumber": bson.M{
					"bsonType":    "string",
					"description": "must be a hex string and is required",
				},
				"updatedAt": bson.M{
					"bsonType":    "string",
					"description": "must be a string and is required",
				},
			},
		},
	}
	ensureCollection(db, "tokenBalances", tokenBalanceValidator)

	// Initialize collections
	initializeCollections(db)

	// Initialize sync state collection
	_, err = db.Collection("sync_state").UpdateOne(
		ctx,
		bson.M{"_id": "last_synced_block"},
		bson.M{"$setOnInsert": bson.M{"block_number": "0x0"}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		Logger.Error("Failed to initialize sync state collection", zap.Error(err))
	}

	return client
}

func ensureCollection(db *mongo.Database, name string, validator bson.M) {
	cmd := bson.D{
		{Key: "collMod", Value: name},
		{Key: "validator", Value: validator},
		{Key: "validationLevel", Value: "strict"},
	}

	err := db.RunCommand(context.Background(), cmd).Err()
	if err != nil {
		// If collection doesn't exist, create it with the validator
		if err.Error() == "not found" {
			opts := options.CreateCollection().SetValidator(validator)
			err = db.CreateCollection(context.Background(), name, opts)
			if err != nil {
				Logger.Warn("Could not create collection with validator",
					zap.String("collection", name),
					zap.Error(err))
			} else {
				Logger.Info("Created collection with validator",
					zap.String("collection", name))
			}
		} else {
			Logger.Warn("Could not set up validator",
				zap.String("collection", name),
				zap.Error(err))
		}
	} else {
		Logger.Info("Updated validator for collection",
			zap.String("collection", name))
	}
}

func initializeCollections(db *mongo.Database) {
	ctx := context.Background()

	// Initialize token balances collection with compound index
	tokenBalancesCollection := db.Collection("tokenBalances")
	_, err := tokenBalancesCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "contractAddress", Value: 1},
				{Key: "holderAddress", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		Logger.Error("Failed to create index for token balances collection", zap.Error(err))
	}

	// Initialize pending token contracts collection with compound index
	pendingTokenContractsCollection := db.Collection("pending_token_contracts")
	_, err = pendingTokenContractsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.D{
				{Key: "contractAddress", Value: 1},
				{Key: "txHash", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		Logger.Error("Failed to create index for pending token contracts collection", zap.Error(err))
	}

	// Also add index on the processed field for efficient querying
	_, err = pendingTokenContractsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "processed", Value: 1}},
			Options: options.Index().SetName("processed_idx"),
		},
	)
	if err != nil {
		Logger.Error("Failed to create processed index for pending token contracts collection", zap.Error(err))
	}

	// Initialize token transfers collection with indexes
	tokenTransfersCollection := db.Collection("tokenTransfers")
	_, err = tokenTransfersCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "contractAddress", Value: 1},
					{Key: "blockNumber", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "from", Value: 1},
					{Key: "blockNumber", Value: 1},
				},
			},
			{
				Keys: bson.D{
					{Key: "to", Value: 1},
					{Key: "blockNumber", Value: 1},
				},
			},
			{
				Keys:    bson.D{{Key: "txHash", Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		},
	)
	if err != nil {
		Logger.Error("Failed to create indexes for token transfers collection", zap.Error(err))
	} else {
		Logger.Info("Token transfers collection initialized with indexes")
	}

	// Initialize CoinGecko collection with empty document
	_, err = db.Collection("coingecko").UpdateOne(
		ctx,
		bson.M{},
		bson.M{"$setOnInsert": bson.M{
			"marketCapUSD": 0.0,
			"priceUSD":     0.0,
			"lastUpdated":  time.Now(),
		}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		Logger.Error("Failed to initialize CoinGecko collection", zap.Error(err))
	}

	// Create and set up the rest of the collections
	ensureCollection(db, "blocks", nil)
	ensureCollection(db, "transfer", nil)
	ensureCollection(db, "validators", nil)
	ensureCollection(db, "contractCode", nil)
	ensureCollection(db, "transactionByAddress", nil)
	ensureCollection(db, "internalTransactionByAddress", nil)
	ensureCollection(db, "contracts", nil)
	ensureCollection(db, "addresses", nil)
	ensureCollection(db, "walletCount", nil)
	ensureCollection(db, "dailyTransactionsVolume", nil)
	ensureCollection(db, "totalCirculatingSupply", nil)
	ensureCollection(db, "sync_state", nil)

	Logger.Info("All collections initialized successfully")
}

// Client instance
var DB *mongo.Client = ConnectDB()

// Getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("qrldata-z").Collection(collectionName)
	return collection
}

// Getter for contracts collection
func GetContractsCollection() *mongo.Collection {
	return GetCollection(DB, CONTRACT_CODE_COLLECTION)
}

// Getter for validator collection
func GetValidatorCollection() *mongo.Collection {
	return GetCollection(DB, VALIDATORS_COLLECTION)
}

// Getter for token balances collection
func GetTokenBalancesCollection() *mongo.Collection {
	return GetCollection(DB, "tokenBalances")
}

// GetTokenTransfersCollection returns the tokenTransfers collection
func GetTokenTransfersCollection() *mongo.Collection {
	// Use GetCollection with explicit collection name
	coll := GetCollection(DB, "tokenTransfers")

	// Log that we're getting a reference to the collection
	Logger.Debug("Getting tokenTransfers collection reference")

	return coll
}

func GetListCollectionNames(client *mongo.Client) []string {
	result, err := client.Database("qrldata-z").ListCollectionNames(
		context.TODO(),
		bson.D{})

	if err != nil {
		log.Fatal(err)
	}

	return result
}
