package sstable

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

var ErrKeyNotFound = errors.New("key not found in sstable")
var ErrInvalidFile = errors.New("invalid sstable file: magic number mismatch")

type SSTableReader struct {
	file        *os.File
	indexEntries []IndexEntry
}

func OpenSSTableReader(filePath string) (*SSTableReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	fileSize := stat.Size()
	if fileSize < 12 {
		file.Close()
		return nil, ErrInvalidFile
	}

	footerBuf := make([]byte, 12)
	if _, err := file.ReadAt(footerBuf, fileSize-12); err != nil {
		file.Close()
		return nil, err
	}

	reader := bytes.NewReader(footerBuf)
	var indexBlockOffset int64
	var magic uint32
	binary.Read(reader, binary.BigEndian, &indexBlockOffset)
	binary.Read(reader, binary.BigEndian, &magic)

	if magic != MagicNumber {
		file.Close()
		return nil, ErrInvalidFile
	}

	indexDataSize := fileSize - 12 - indexBlockOffset
	indexData := make([]byte, indexDataSize)
	if _, err := file.ReadAt(indexData, indexBlockOffset); err != nil {
		file.Close()
		return nil, err
	}

	var indexEntries []IndexEntry
	offsetTracker := 0
	for offsetTracker < len(indexData) {
		entry, bytesConsumed, err := DeserializeIndexEntry(indexData[offsetTracker:])
		if err != nil {
			break
		}
		indexEntries = append(indexEntries, entry)
		offsetTracker += bytesConsumed
	}

	return &SSTableReader{
		file:        file,
		indexEntries: indexEntries,
	}, nil
}

func (r *SSTableReader) Get(targetKey []byte) ([]byte, error) {
	low := 0
	high := len(r.indexEntries) - 1
	targetBlockOffset := int64(-1)

	for low <= high {
		mid := (low + high) / 2
		comparison := bytes.Compare(r.indexEntries[mid].Key, targetKey)

		if comparison >= 0 {
			targetBlockOffset = r.indexEntries[mid].Offset
			high = mid - 1 
		} else {
			low = mid + 1
		}
	}

	if targetBlockOffset == -1 && len(r.indexEntries) > 0 {
		targetBlockOffset = r.indexEntries[len(r.indexEntries)-1].Offset
	} else if targetBlockOffset == -1 {
		return nil, ErrKeyNotFound
	}

	stat, _ := r.file.Stat()
	maxReadSize := stat.Size() - 12 - targetBlockOffset 

	buf := make([]byte, maxReadSize)
	if _, err := r.file.ReadAt(buf, targetBlockOffset); err != nil {
		return nil, err
	}

	bufReader := bytes.NewReader(buf)
	for bufReader.Len() > 0 {
		var kLen, vLen uint32
		if err := binary.Read(bufReader, binary.BigEndian, &kLen); err != nil {
			break
		}
		if err := binary.Read(bufReader, binary.BigEndian, &vLen); err != nil {
			break
		}

		k := make([]byte, kLen)
		v := make([]byte, vLen)
		bufReader.Read(k)
		bufReader.Read(v)

		if bytes.Equal(k, targetKey) {
			return v, nil
		}
	}

	return nil, ErrKeyNotFound
}

func (r *SSTableReader) Close() error {
	return r.file.Close()
}