package db

/*
TODO: Implement the customDb in this file.

You can change the file name if you want.

Make sure your db implement the DbIf interface in db.go.

Since the customDb will not test in TA's environment, you can just harcode the db's configuration at this file, e.g. db address/keys.
*/
import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type mongoDb struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type mongoRecord struct {
	ID    string `bson:"_id"`
	Value string `bson:"value"`
}

func newMongoDb() (*mongoDb, error) {
	client, err := mongo.Connect(
		options.Client().ApplyURI("mongodb://127.0.0.1:27018"),
	)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return &mongoDb{
		client:     client,
		collection: client.Database("cndi_pj2").Collection("services"),
	}, nil
}

func (m *mongoDb) Release() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.client.Disconnect(ctx)
}

func (m *mongoDb) Save(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": key}
	update := bson.M{
		"$set": bson.M{"value": value},
	}

	_, err := m.collection.UpdateOne(
		ctx,
		filter,
		update,
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

func (m *mongoDb) Exist(key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := m.collection.CountDocuments(
		ctx,
		bson.M{"_id": key},
	)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (m *mongoDb) Load(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var record mongoRecord

	err := m.collection.FindOne(
		ctx,
		bson.M{"_id": key},
	).Decode(&record)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return record.Value, nil
}

func (m *mongoDb) Remove(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.DeleteOne(
		ctx,
		bson.M{"_id": key},
	)
	return err
}

func (m *mongoDb) SaveAll(data map[string]string) error {
	for key, value := range data {
		if err := m.Save(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (m *mongoDb) LoadAll() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() {
	    _ = cursor.Close(context.Background())
	}()

	result := make(map[string]string)

	for cursor.Next(ctx) {
		var record mongoRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		result[record.ID] = record.Value
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (m *mongoDb) RemoveAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.collection.DeleteMany(ctx, bson.M{})
	return err
}
