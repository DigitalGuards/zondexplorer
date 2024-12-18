package db

import (
	"QRLtoMongoDB-PoS/configs"
	"QRLtoMongoDB-PoS/models"
	"QRLtoMongoDB-PoS/rpc"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func InsertValidators(validators models.ResultValidator) {
	filter := primitive.D{{Key: "jsonrpc", Value: 2.0}}
	update := primitive.D{
		{Key: "$set", Value: primitive.D{
			{Key: "resultvalidator", Value: validators},
		}},
	}
	opts := options.Update().SetUpsert(true)
	_, err := configs.ValidatorsCollections.UpdateOne(context.TODO(), filter, update, opts)
	if err != nil {
		configs.Logger.Warn("Failed to insert in the validators collection: ", zap.Error(err))
	}
}

func UpdateValidators(sum uint64, previousHash string) {
	currentEpoch := int(sum) / 100
	parentEpoch := int(GetBlockNumberFromHash(previousHash)) / 100
	if currentEpoch != parentEpoch {
		validators := rpc.GetValidators()
		InsertValidators(validators)
		fmt.Println("Succesfully updated validators")
	}
}
