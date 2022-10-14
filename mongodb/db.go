package mongodb

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	AuthSource string
	Username   string
	Password   string
	Opts       string
	Database   string
	Hosts      []string
}

type Client struct {
	mclient *mongo.Client
	db      *mongo.Database
}

// NewClient method takes a config map argument
func NewClient(conf Config, connnectionStr string) (*Client, error) {
	var client = &Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mclient, err := mongo.Connect(ctx, options.Client().ApplyURI(connnectionStr))
	if err != nil {
		return nil, err
	}
	client.mclient = mclient
	client.db = mclient.Database(conf.Database)
	return client, nil
}

func (c *Client) GetDb() *mongo.Database {
	return c.db
}

func (c *Client) Ping() error {
	return c.mclient.Ping(context.TODO(), nil)
}

func (c *Client) GenerateID() primitive.ObjectID {
	return primitive.NewObjectID()
}
