package main

import (
	"fmt"
	"log"

	"github.com/iamsalnikov/boltmigration"
	"go.etcd.io/bbolt"
)

func main() {
	db, err := bbolt.Open("db", 0666, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	// Add migration
	boltmigration.Add("0001_init", func(db *bbolt.DB) error {
		return db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte("waat"))
			return err
		})
	})

	boltmigration.SetDatabase(db)

	newMigs, err := boltmigration.NewMigrationNames()
	if err != nil {
		log.Fatalln(err)
	}

	for _, m := range newMigs {
		fmt.Println(m)
	}

	boltmigration.Apply()
}
