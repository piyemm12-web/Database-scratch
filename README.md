# Database-scratch

A high-performance LSM-tree based Key-Value Database Store written from scratch in Go.

## Features
- **MemTable**: Fast in-memory key-value storage backed by a Skip List structure.
- **WAL (Write-Ahead Logging)**: Ensures durability and crash recovery for all write operations.
- **SSTable**: Sorted String Table binary file structures with sparse indexing for efficient disk-based reads.
- **Compaction**: Automated background compaction routines to merge sorted string tables, eliminate stale keys, and reclaim disk space.
- **KV Server**: Dedicated server executable and RESTful API endpoints for key-value store interactions and client communication.

## Project Structure
- `cmd/kv-server`: Main entry point and server executable for the key-value database store.
- `internal/`: Core storage engine modules including memtable, wal, sstable, compaction, and core database coordination logic.
