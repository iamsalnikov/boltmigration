package migration

import (
	"encoding/json"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

// AppliedFunc returns list of applied migrations
type AppliedFunc func(db *bbolt.DB) ([]string, error)

// MarkAppliedFunc marks migration as applied
type MarkAppliedFunc func(db *bbolt.DB, name string) error

// UpFunc applies migrations
type UpFunc func(db *bbolt.DB) error

type mig struct {
	name string
	up   UpFunc
}

type migrationRecord struct {
	Name        string
	AppliedTime time.Time
}

func (mr migrationRecord) key() []byte {
	return []byte(mr.Name)
}

const migrationsTable = "migrations"

var defaultAppliedFunc = AppliedFunc(func(db *bbolt.DB) ([]string, error) {
	createMigrationsBucket(db)

	applied := make([]string, 0)

	err := db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(migrationsTable))

		return bucket.ForEach(func(k, v []byte) error {
			mig := &migrationRecord{}
			err := json.Unmarshal(v, mig)
			if err != nil {
				return err
			}

			applied = append(applied, mig.Name)

			return nil
		})
	})

	return applied, err
})

var defaultMarkAppliedFunc = MarkAppliedFunc(func(db *bbolt.DB, name string) error {
	createMigrationsBucket(db)
	mr := migrationRecord{
		Name:        name,
		AppliedTime: time.Now(),
	}

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(migrationsTable))

		data, err := json.Marshal(&mr)
		if err != nil {
			return err
		}

		return b.Put(mr.key(), data)
	})
})

var migrations = make(map[string]mig)
var db *bbolt.DB
var getApplied = defaultAppliedFunc
var markApplied = defaultMarkAppliedFunc

func resetMigrations() {
	migrations = make(map[string]mig)
}

func resetAppliedFunc() {
	getApplied = defaultAppliedFunc
}

func resetMarkAppliedFunc() {
	markApplied = defaultMarkAppliedFunc
}

func createMigrationsBucket(db *bbolt.DB) error {
	return db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(migrationsTable))
		return err
	})
}

// Add adds mig to queue
// Use this function in init()
func Add(name string, up UpFunc) {
	migrations[name] = mig{
		name: name,
		up:   up,
	}
}

// SetDatabase sets a database that we should use for applying migrations
func SetDatabase(database *bbolt.DB) {
	db = database
}

// NewMigrationNames returns names of new migrations
func NewMigrationNames() ([]string, error) {
	appliedNames, err := getApplied(db)
	if err != nil {
		return []string{}, err
	}

	applied := map[string]bool{}
	for _, name := range appliedNames {
		applied[name] = true
	}

	result := make([]string, 0)
	for name := range migrations {
		if !applied[name] {
			result = append(result, name)
		}
	}

	sort.Strings(result)

	return result, nil
}

// Apply func applies migrations
func Apply() error {
	newNames, err := NewMigrationNames()
	if err != nil {
		return err
	}

	for _, name := range newNames {
		err = migrations[name].up(db)
		if err != nil {
			return err
		}

		err = markApplied(db, name)
		if err != nil {
			return err
		}
	}

	return nil
}
