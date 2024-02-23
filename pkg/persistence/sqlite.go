package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
    "io/fs"
    "os"
    "path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valentin-carl/stattrack/pkg/measurements"
)

type SqliteBackend struct {
	ctx    context.Context
	values <-chan measurements.Measurement
	mType  measurements.MeasurementType
	db     *sql.DB
}

//
// Model: 1 DB, 3 backends => makes it easier to integrate with other code in cmd/main.go
//

func NewSqliteBackend(
	ctx context.Context,
	values <-chan measurements.Measurement,
    outdir string,
	mType measurements.MeasurementType,
	dbFilename string,
) (*SqliteBackend, error) {

	log.Println("creating new sqlite backend")

    // create directory to put DB into
	err := os.MkdirAll(outdir, fs.ModePerm)
	if err != nil {
		log.Println("error occurred while trying to create output directory", err.Error())
		return nil, err
	}

    // create db path from outdir & dbFilename
	dbPath := path.Join(outdir, dbFilename)
    log.Println("dbPath:", dbPath)

	DB, err := getDB(ctx, dbPath)
	if err != nil {
		log.Println("could not open database")
		return nil, err
	}

	b := &SqliteBackend{
		ctx:    ctx,
		values: values,
		mType:  mType,
		db:     DB,
	}

	// create tables
	query := createTable[mType]
	_, err = b.db.ExecContext(ctx, query)
	if err != nil {
		log.Println("something went wrong while trying to create a table | query:", query)
		return nil, err
	}

	return b, nil
}

func (b *SqliteBackend) Start() error {

	log.Printf("sqlite backend for %d starting\n", b.mType)

	//defer b.db.Close()

	var err error

	for {
		select {
		case value := <-b.values:
			{
                // todo remove logging statement later
                log.Println("inserting value into db")
                err = insertValue(b.ctx, value, b.db)
                if err != nil {
                    log.Println("sqlite backend: error while trying to insert values into DB")
                    b.db.Close()
                    return err
                }
			}
		case <-b.ctx.Done():
			{
				log.Println("sqlite backend: context cancelled, quitting ...")
				goto TheEnd
			}
		}
	}

TheEnd:
	log.Println("sqlite backend done")
    b.db.Close()

	return err
}

//
// DATABASE HELPER CODE
//

func getDB(ctx context.Context, dbPath string) (*sql.DB, error) {

	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		log.Println("error creating database file")
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		log.Println("could not establish database connection")
		return nil, err
	}

	return db, nil
}

var createTable = map[measurements.MeasurementType]string{
	0: `CREATE TABLE cpu (
    timestamp INTEGER,
    user INTEGER,
    system INTEGER,
    idle INTEGER,
    nice INTEGER,
    total INTEGER,
    userp FLOAT,
    systemp FLOAT,
    idlep FLOAT
);`,
	1: `CREATE TABLE memory (
    timestamp INTEGER,
    free INTEGER,
    total INTEGER,
    active INTEGER,
    cached INTEGER,
    inactive INTEGER,
    swapFree INTEGER,
    swapTotal INTEGER,
    swapUsed INTEGER,
    used INTEGER,
    freep FLOAT
);`,
	2: `CREATE TABLE network (
    timestamp INTEGER,
    name TINYTEXT,
    RxBytes INTEGER,
    TxBytes INTEGER
);`,
}

func insertValue(ctx context.Context, value measurements.Measurement, db *sql.DB) error {

	var (
		err error
		t   measurements.MeasurementType
	)

	switch value.(type) {
	case measurements.CPUMeasurement:
		t = 0
	case measurements.MemoryMeasurement:
		t = 1
	case measurements.NetworkMeasurement:
		t = 2
	}

	query := fmt.Sprintf(insert[t], strings.Join(value.Record(), ", "))

	transaction, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Println("could not open new transaction")
		return err
	}

	_, err = transaction.ExecContext(ctx, query)
	if err != nil {
		log.Println("error while executing insert statement")
		return err
	}

	err = transaction.Commit()
	if err != nil {
		log.Println("error while commiting transaction")
		return err
	}

	return err
}

var insert = map[measurements.MeasurementType]string{
	0: `INSERT INTRO cpu (
    timestamp,
    user,
    system,
    idle,
    nice,
    total,
    userp,
    systemp,
    idlep
) values (
    %s
);`,
	1: `INSERT INTO memory (
    timestamp,
    free,
    total,
    active,
    cached,
    inactive,
    swapFree,
    swapTotal,
    swapUsed,
    used,
    freep
) values (
    %s
);`,
	2: `INSERT INTO network (
    timestamp,
    name,
    RxBytes,
    TxBytes
) values (
    %s
);`,
}
