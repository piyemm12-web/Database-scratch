package sstable

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidIndexEntry = errors.New("sstable index entry is malformed or truncated")
)

type IndexEntry struct {
	Key    []byte
	Offset int64 
}

func SerializeIndexEntry(entry IndexEntry) ([]byte, error) {
	buf := new(bytes.Buffer)
	keyLen := uint32(len(entry.Key))

	if err := binary.Write(buf, binary.BigEndian, keyLen); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, entry.Offset); err != nil {
		return nil, err
	}

	if _, err := buf.Write(entry.Key); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DeserializeIndexEntry(data []byte) (IndexEntry, int, error) {
	if len(data) < 12 {
		return IndexEntry{}, 0, ErrInvalidIndexEntry
	}

	reader := bytes.NewReader(data)
	var keyLen uint32
	var offset int64

	if err := binary.Read(reader, binary.BigEndian, &keyLen); err != nil {
		return IndexEntry{}, 0, err
	}
	if err := binary.Read(reader, binary.BigEndian, &offset); err != nil {
		return IndexEntry{}, 0, err
	}

	if int64(len(data)) < 12+int64(keyLen) {
		return IndexEntry{}, 0, ErrInvalidIndexEntry
	}

	key := make([]byte, keyLen)
	if _, err := reader.Read(key); err != nil {
		return IndexEntry{}, 0, err
	}

	totalBytesRead := 12 + int(keyLen)
	return IndexEntry{Key: key, Offset: offset}, totalBytesRead, nil
}