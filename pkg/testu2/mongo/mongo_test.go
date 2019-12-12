package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongo(t *testing.T) {
	m := New()
	t.Logf("mongo host: %s", m.Host)

	s, err := mongo.Connect(context.Background(),
		options.Client().ApplyURI(fmt.Sprintf("mongodb://%s/test", m.Host)))
	assert.Nil(t, err)
	err = s.Ping(context.Background(), nil)
	assert.Nil(t, err)
	c := s.Database("pero").Collection("zero")

	_, err = c.InsertOne(context.Background(), bson.D{{"_id", "0"}})
	assert.Nil(t, err)

	m.Stop()
}
