package sstable

import (
	"bytes"
	"encoding/binary"
	"os"
)

const MagicNumber uint32 = 0x48564B53 

type SSTableBuilder struct {
	file        *os.File
	currentOffset int64
	indexEntries []IndexEntry
}

func NewSSTableBuilder(filePath string) (*SSTableBuilder, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	return &SSTableBuilder{file: file}, nil
}

func (b *SSTableBuilder) Add(key, value []byte) error {
	blockOffset := b.currentOffset

	buf := new(bytes.Buffer)
	keyLen := uint32(len(key))
	valLen := uint32(len(value))

	binary.Write(buf, binary.BigEndian, keyLen)
	binary.Write(buf, binary.BigEndian, valLen)
	buf.Write(key)
	buf.Write(value)

	data := buf.Bytes()
	if _, err := b.file.Write(data); err != nil {
		return err
	}

	b.indexEntries = append(b.indexEntries, IndexEntry{
		Key:    key,
		Offset: blockOffset,
	})

	b.currentOffset += int64(len(data))
	return nil
}

func (b *SSTableBuilder) Finalize() error {
	// 1. Record where the Index Block starts
	indexBlockOffset := b.currentOffset

	for _, entry := range b.indexEntries {
		serializedIndex, err := SerializeIndexEntry(entry)
		if err != nil {
			return err
		}
		if _, err := b.file.Write(serializedIndex); err != nil {
			return err
		}
		b.currentOffset += int64(len(serializedIndex))
	}

	footerBuf := new(bytes.Buffer)
	if err := binary.Write(footerBuf, binary.BigEndian, indexBlockOffset); err != nil {
		return err
	}
	if err := binary.Write(footerBuf, binary.BigEndian, MagicNumber); err != nil {
		return err
	}

	if _, err := b.file.Write(footerBuf.Bytes()); err != nil {
		return err
	}

	return b.file.Close()
}