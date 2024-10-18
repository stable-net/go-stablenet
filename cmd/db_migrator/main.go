//go:build db_migrator

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/linxGnu/grocksdb"
)

var (
	srcDataDir = flag.String("src", "", "datadir for go-wemix")
	dstDataDir = flag.String("dst", "", "datadir for go-wemix-qbft")
	batchSize  = flag.Uint("batch", 50*1024*1024, "Key-value size (Byte) to batch for chaindata")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "-src <datadir (go-wemix)> -dst <datadir (go-wemix-qbft)> [-batch <batch_size>]")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Migrate chaindata for go-wemix to chaindata for go-wemix-qbft.`)
	}
}

func main() {
	flag.Parse()

	if *srcDataDir == "" {
		panic(fmt.Errorf("empty src"))
	}

	if *dstDataDir == "" {
		panic(fmt.Errorf("empty dst"))
	}

	if *batchSize == 0 {
		panic(fmt.Errorf("batch count must be positive number"))
	}

	migrateChaindata()
}

func migrateChaindata() {
	srcChaindataDir := *srcDataDir + "/geth/chaindata"
	dstChaindataDir := *dstDataDir + "/geth/chaindata"

	fmt.Println("Migrating chaindata", srcChaindataDir, "to", dstChaindataDir)

	// Open go-wemix chaindata
	opts := grocksdb.NewDefaultOptions()
	db, err := grocksdb.OpenDb(opts, srcChaindataDir)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Open go-wemix-qbft chaindata
	newDb, err := rawdb.Open(rawdb.OpenOptions{
		Type:      "pebble",
		Directory: dstChaindataDir,
		Namespace: "",
		Cache:     0,
		Handles:   0,
		ReadOnly:  false,
	})
	if err != nil {
		panic(err)
	}
	defer newDb.Close()

	ro := grocksdb.NewDefaultReadOptions()
	it := db.NewIterator(ro)
	defer it.Close()

	batch := newDb.NewBatch()

	// Iterate all the KV pairs
	count := 0
	for it.SeekToFirst(); it.Valid(); it.Next() {
		key := it.Key().Data()
		value := it.Value().Data()
		if err := batch.Put(key, value); err != nil {
			panic(err)
		}

		//fmt.Println("key: %s, value: %s", hex.EncodeToString(key), hex.EncodeToString(value))

		count++

		if uint(batch.ValueSize()) >= *batchSize {
			if err := batch.Write(); err != nil {
				panic(err)
			}

			fmt.Println("Write", count, "pairs", batch.ValueSize(), "bytes")

			count = 0
			batch.Reset()
		}
	}

	if err := it.Err(); err != nil {
		panic(err)
	}

	// Commit the rest chaindata
	if err := batch.Write(); err != nil {
		panic(err)
	}
	fmt.Println("Write", count, "pairs", batch.ValueSize(), "bytes")

	fmt.Println("Done")
}
