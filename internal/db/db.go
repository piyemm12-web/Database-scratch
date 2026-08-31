package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"my-kv-store/internal/compaction"
	"my-kv-store/internal/memtable"
	"my-kv-store/internal/sstable"
	"my-kv-store/internal/wal"
)

var ErrKeyNotFound = errors.New("key not found")

type DB struct {
	opts       Options
	wal        *wal.WAL
	memtable   *memtable.Memtable
	immutable  *memtable.Memtable 
	sstFiles   []string
	mu         sync.RWMutex
	compaction *compaction.CompactionManager
}

func Open(opts Options) (*DB, error) {
	if err := os.MkdirAll(opts.DataDir, 0755); err != nil {
		return nil, err
	}

	walPath := filepath.Join(opts.DataDir, "wal.log")
	walInstance, err := wal.OpenWAL(walPath)
	if err != nil {
		return nil, err
	}

	activeMemtable := memtable.NewMemtable(opts.MaxMemtableSize)

	db := &DB{
		opts:       opts,
		wal:        walInstance,
		memtable:   activeMemtable,
		compaction: compaction.NewCompactionManager(opts.DataDir),
	}

	if err := db.recoverWAL(); err != nil {
		return nil, err
	}

	if err := db.loadSSTables(); err != nil {
		return nil, err
	}

	db.compaction.Start(
		func() []string {
			db.mu.RLock()
			defer db.mu.RUnlock()
			return db.sstFiles
		},
		func(newFiles []string) {
			db.mu.Lock()
			defer db.mu.Unlock()
			db.sstFiles = newFiles
		},
	)

	return db, nil
}

func (db *DB) Put(key, value []byte) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	entry := &wal.Entry{Key: key, Value: value}
	if err := db.wal.Append(entry); err != nil {
		return err
	}

	db.memtable.Put(key, value)
	if db.memtable.IsFull() {
		if err := db.flushMemtableLocked(); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Get(key []byte) ([]byte, error) {
	db.mu.RLock()
	if val, found := db.memtable.Get(key); found {
		db.mu.RUnlock()
		if len(val) == 0 {
			return nil, ErrKeyNotFound 
		}
		return val, nil
	}

	if db.immutable != nil {
		if val, found := db.immutable.Get(key); found {
			db.mu.RUnlock()
			if len(val) == 0 {
				return nil, ErrKeyNotFound
			}
			return val, nil
		}
	}

	sstFilesCopy := make([]string, len(db.sstFiles))
	copy(sstFilesCopy, db.sstFiles)
	db.mu.RUnlock()

	for i := len(sstFilesCopy) - 1; i >= 0; i-- {
		reader, err := sstable.OpenSSTableReader(sstFilesCopy[i])
		if err != nil {
			continue
		}
		val, err := reader.Get(key)
		reader.Close()
		if err == nil {
			return val, nil
		}
	}

	return nil, ErrKeyNotFound
}

func (db *DB) Delete(key []byte) error {
	return db.Put(key, []byte{})
}

func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.compaction.Stop()
	return db.wal.Close()
}

func (db *DB) recoverWAL() error {
	entries, err := db.wal.ReadAll()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		db.memtable.Put(entry.Key, entry.Value)
	}
	return nil
}

func (db *DB) loadSSTables() error {
	matches, err := filepath.Glob(filepath.Join(db.opts.DataDir, "*.db"))
	if err != nil {
		return err
	}
	db.sstFiles = matches
	return nil
}

func (db *DB) flushMemtableLocked() error {
	entries := db.memtable.Entries()
	if len(entries) == 0 {
		return nil
	}

	sstablePath := filepath.Join(db.opts.DataDir, fmt.Sprintf("sstable_%d.db", len(db.sstFiles)+1))
	builder, err := sstable.NewSSTableBuilder(sstablePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := builder.Add(entry.Key, entry.Value); err != nil {
			builder.Finalize()
			return err
		}
	}

	if err := builder.Finalize(); err != nil {
		return err
	}

	db.sstFiles = append(db.sstFiles, sstablePath)
	db.memtable = memtable.NewMemtable(db.opts.MaxMemtableSize) 

	db.compaction.Trigger()
	return nil
}