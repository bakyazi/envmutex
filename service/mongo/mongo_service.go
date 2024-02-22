package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/bakyazi/envmutex/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoServiceImpl struct {
	opts       *options.ClientOptions
	db         string
	collection string
}

func (m *MongoServiceImpl) GetEnvironments() ([]model.Environment, error) {
	var envs []model.Environment
	err := m.runCommand(func(ctx context.Context, client *mongo.Client) error {
		col := client.Database(m.db).Collection(m.collection)
		curr, err := col.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"name", 1}}))
		if err != nil {
			return err
		}
		err = curr.All(ctx, &envs)
		if err != nil {
			return err
		}
		return nil
	})
	return envs, err
}

func (m *MongoServiceImpl) LockEnvironment(name, owner string) error {
	return m.runCommand(func(ctx context.Context, client *mongo.Client) error {
		col := client.Database(m.db).Collection(m.collection)
		result, err := col.UpdateOne(ctx, bson.D{{"name", name}, {"status", "Free"}},
			bson.D{{"$set", bson.D{{"status", "Locked"}, {"owner", owner}, {"date", time.Now()}}}})
		if err != nil {
			return err
		}

		if result.ModifiedCount == 0 {
			return errors.New("cannot lock environment")
		}
		return nil
	})
}

func (m *MongoServiceImpl) ReleaseEnvironment(name, owner string) error {
	return m.runCommand(func(ctx context.Context, client *mongo.Client) error {
		col := client.Database(m.db).Collection(m.collection)
		result, err := col.UpdateOne(ctx, bson.D{{"name", name}, {"owner", owner}, {"status", "Locked"}},
			bson.D{{"$set", bson.D{{"status", "Free"}, {"owner", ""}, {"date", time.Now()}}}})
		if err != nil {
			return err
		}

		if result.ModifiedCount == 0 {
			return errors.New("cannot lock environment")
		}
		return nil
	})
}

func NewMongoService(username, password, host, appName, db, collection string) *MongoServiceImpl {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=%s", username, password, host, appName)).SetServerAPIOptions(serverAPI)
	return &MongoServiceImpl{
		opts:       opts,
		db:         db,
		collection: collection,
	}
}

type commandFunc func(ctx context.Context, client *mongo.Client) error

func (m *MongoServiceImpl) runCommand(fn commandFunc) (err error) {
	var client *mongo.Client
	client, err = mongo.Connect(context.TODO(), m.opts)
	if err != nil {
		return
	}
	defer func() {
		err = client.Disconnect(context.TODO())
	}()

	return fn(context.TODO(), client)
}
