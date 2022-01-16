package db

import (
	"auth-api/config"
	"fmt"

	"github.com/couchbase/gocb/v2"
)

func CouchbaseConnect(c *config.Config) *gocb.Cluster {
	opts := gocb.ClusterOptions{
		Username: c.Couchbase.CBUsername,
		Password: c.Couchbase.CBPassword,
	}
	cluster, err := gocb.Connect(fmt.Sprintf("couchbase://%s", c.Couchbase.CBAddress), opts)
	if err != nil {
		fmt.Println("Couchbase connection error")
		panic(err)
	}
	cluster.QueryIndexes().CreatePrimaryIndex("Users", &gocb.CreatePrimaryQueryIndexOptions{IgnoreIfExists: true})
	return cluster
}
