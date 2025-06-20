Migrate chaindata for `gwemix` to chaindata for `geth`.

**First of all, check which db engine is used to manage `gwemix` chaindata.**

```shell
ls <path-to-gwemix-datadir>/geth/chaindata
```

`gwemix` manages chaindata with file extension either `*.sst` or `*.ldb`.
If LevelDB (`*.ldb`) is used, just run `geth` with `--db.engine leveldb` option added.

```shell
geth \
  --db.engine leveldb \
  --datadir <path-to-gwemix-datadir>/geth/chaindata \
  ...
```

The below migration guide is for chaindata managed by RocksDB (`*.sst`) to migrate into Pebble.

## Prerequisites

Secure enough free storage space.

Migrating chaindata *including* ancient chaindata requires as much free space as datadir occupied. (**recommended**)

Migrating chaindata *excluding* ancient chaindata requires as much free space as datadir occupied excluding ancient chaindata.
After migration, you need to move ancient chaindata into migrated datadir, or run `geth` with `--datadir.ancient <path-to-gwemix-ancient-chaindata>` option added.

## Preparing migration tool

Setup environment variables for preparing migration tool `db_migrator`.

```shell
export GWEMIX_REPO=<path-to-go-wemix-repo>
export GWEMIX_WBFT_REPO=<path-to-go-wemix-wbft-repo>

export GWEMIX_DATADIR=<path-to-gwemix-datadir>
export GWEMIX_WBFT_DATADIR=<path-to-geth-datadir> # Create new directory
```

### Getting migration tool from releases (**recommended**)

Get `db_migrator` from releases and locate it to `$GWEMIX_WBFT_REPO`.

### Building migration tool from the source

If rocksdb is not built, build it first from go-wemix repo.

```shell
cd $GWEMIX_REPO
make rocksdb
```

Build `db_migrator`.

```shell
sudo apt install -y libjemalloc-dev liblz4-dev libsnappy-dev libzstd-dev libudev-dev zlib1g-dev

cd $GWEMIX_WBFT_REPO
CGO_CFLAGS="-I$GWEMIX_REPO/rocksdb/include" CGO_LDFLAGS="-L$GWEMIX_REPO/rocksdb -lm -lstdc++ -lpthread -lrt -ldl -lsnappy -llz4 -lzstd -ljemalloc" go build -tags db_migrator ./cmd/db_migrator
```

Note that `gwemix` v0.10.9 uses RocksDB v6.28.2 and the corresponding grocksdb (RocksDB wrapper) v1.6.46.
If `gwemix` uses other RocksDB version, adjusting grocksdb version might be needed.

## Migration including ancient chaindata

Before migration, stop running gwemix to prevent DB inconsistent issues.

Get current chaindata info.

```shell
gwemix db inspect --datadir $GWEMIX_DATADIR --syncmode snap
gwemix db metadata --datadir $GWEMIX_DATADIR --syncmode snap
```

Copy datadir containing node info, keystores, and chaindata to setup datadir for `geth`.

```shell
cp -a $GWEMIX_DATADIR $GWEMIX_WBFT_DATADIR
```

Keep ancient chaindata then remove all the rest chaindata.

```shell
mv $GWEMIX_WBFT_DATADIR/geth/chaindata/ancient $GWEMIX_WBFT_DATADIR/geth
rm -r $GWEMIX_WBFT_DATADIR/geth/chaindata
```

Migrate chaindata by using `db_migrator`.

```shell
$GWEMIX_WBFT_REPO/db_migrator -src $GWEMIX_DATADIR -dst $GWEMIX_WBFT_DATADIR
```

Restore ancient chaindata.

```shell
mv $GWEMIX_WBFT_DATADIR/geth/ancient $GWEMIX_WBFT_DATADIR/geth/chaindata
```

Check migrated chaindata if it is same info for `$GWEMIX_DATADIR`.

```shell
geth db inspect --datadir $GWEMIX_WBFT_DATADIR --syncmode snap
geth db metadata --datadir $GWEMIX_WBFT_DATADIR --syncmode snap
```

Migration has done.

Run `geth` with `--datadir $GWEMIX_WBFT_DATADIR` option added.
The default value for `--db.engine` is `pebble` so it can be omitted.

```shell
geth \
  --datadir $GWEMIX_WBFT_DATADIR \
  ...
```
