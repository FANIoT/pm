package grifts

import (
	"log"

	"github.com/gobuffalo/envy"
	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"

	grift "github.com/markbates/grift/grift"
)

var _ = grift.Desc("mongo", "Creates mongo database and collections")
var _ = grift.Add("mongo", func(c *grift.Context) error {
	// Create mongodb connection
	url := envy.Get("DB_URL", "mongodb://127.0.0.1")
	client, err := mgo.NewClient(url)
	if err != nil {
		return err
	}
	if err := client.Connect(c); err != nil {
		return err
	}
	log.Printf("DB url: %s\n", url)

	db := client.Database("i1820")

	// projects collection
	cp := db.Collection("things")
	pnames, err := cp.Indexes().CreateMany(
		c,
		[]mgo.IndexModel{
			mgo.IndexModel{
				Keys: bson.NewDocument(
					bson.EC.String("perimeter", "2dsphere"),
				),
			},
		},
	)
	if err != nil {
		return err
	}
	log.Printf("DB [projects] index: %s\n", pnames)

	// things collection
	ct := db.Collection("things")
	tnames, err := ct.Indexes().CreateMany(
		c,
		[]mgo.IndexModel{
			mgo.IndexModel{
				Keys: bson.NewDocument(
					bson.EC.Int32("project", 1),
				),
			},
			mgo.IndexModel{
				Keys: bson.NewDocument(
					bson.EC.String("location", "2dsphere"),
				),
			},
		},
	)
	if err != nil {
		return err
	}
	log.Printf("DB [things] index: %s\n", tnames)

	return nil
})
