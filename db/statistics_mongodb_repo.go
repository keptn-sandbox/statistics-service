package db

import (
	"context"
	"github.com/globalsign/mgo/bson"
	"github.com/keptn-sandbox/statistics-service/operations"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const keptnStatsCollection = "keptn-stats"

type StatisticsMongoDBRepo struct {
	DbConnection MongoDBConnection
}

// GetStatistics godoc
func (s StatisticsMongoDBRepo) GetStatistics(from, to time.Time) ([]operations.Statistics, error) {
	collection, err := s.getCollection()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sortOptions := options.Find().SetSort(bson.D{{"from", 1}})
	searchOptions := bson.M{}

	searchOptions["from"] = bson.M{
		"$gt": from,
	}
	searchOptions["to"] = bson.M{
		"$lt": to,
	}

	cur, err := collection.Find(ctx, searchOptions, sortOptions)
	if err != nil {
		return nil, err
	}

	result := []operations.Statistics{}
	defer cur.Close(ctx)
	if cur.RemainingBatchLength() == 0 {
		return nil, NoStatisticsFoundError
	}
	for cur.Next(ctx) {
		stats := &operations.Statistics{}
		err := cur.Decode(stats)
		if err != nil {
			return nil, err
		}

		result = append(result, *stats)
	}

	return result, nil
}

// StoreStatistics godoc
func (s StatisticsMongoDBRepo) StoreStatistics(statistics operations.Statistics) error {
	collection, err := s.getCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, statistics)
	if err != nil {
		return err
	}

	return nil
}

// DeleteStatistics godoc
func (s StatisticsMongoDBRepo) DeleteStatistics(from, to time.Time) error {
	collection, err := s.getCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searchOptions := bson.M{}

	searchOptions["from"] = bson.M{
		"$gt": from.String(),
	}
	searchOptions["to"] = bson.M{
		"$lt": to.String(),
	}

	_, err = collection.DeleteMany(ctx, searchOptions)
	return err
}

func (s StatisticsMongoDBRepo) getCollection() (*mongo.Collection, error) {
	err := s.DbConnection.EnsureDBConnection()
	if err != nil {
		return nil, err
	}

	collection := s.DbConnection.Client.Database(databaseName).Collection(keptnStatsCollection)
	return collection, nil
}
