# Database-scratch

A LSM-tree based Key-Value Store written from scratch in Go (`my-kv-store`).

## Features
- **MemTable**: In-memory key-value storage.
- **WAL (Write-Ahead Logging)**: Crash recovery and persistence.
- **SSTable**: Sorted String Table for disk storage.
- **Compaction**: Background merge and compaction of SSTables.
- **KV Server**: Server executable for key-value store interactions.

## Project Structure
- `cmd/kv-server`: Main entry point for the key-value store server.
- `internal/`: Core modules including memtable, wal, sstable, compaction, and db logic.
