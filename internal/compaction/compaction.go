package compaction

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"my-kv-store/internal/sstable"
)

type CompactionManager struct {
	dataDir     string
	triggerChan chan struct{}
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

func NewCompactionManager(dataDir string) *CompactionManager {
	return &CompactionManager{
		dataDir:     dataDir,
		triggerChan: make(chan struct{}, 1), 
		stopChan:    make(chan struct{}),
	}
}

func (cm *CompactionManager) Start(getSSTables func() []string, updateSSTables func([]string)) {
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		for {
			select {
			case <-cm.triggerChan:
				cm.runCompaction(getSSTables, updateSSTables)
			case <-cm.stopChan:
				return
			}
		}
	}()
}

func (cm *CompactionManager) Trigger() {
	select {
	case cm.triggerChan <- struct{}{}:
	default:
	}
}

func (cm *CompactionManager) Stop() {
	close(cm.stopChan)
	cm.wg.Wait()
}

func (cm *CompactionManager) runCompaction(getSSTables func() []string, updateSSTables func([]string)) {
	sstFiles := getSSTables()
	if len(sstFiles) < 2 {
		return 
	}

	log.Printf("[Compaction] Starting compaction across %d SSTables...", len(sstFiles))

	keyMap := make(map[string][]byte)

	for _, filePath := range sstFiles {
		reader, err := sstable.OpenSSTableReader(filePath)
		if err != nil {
			continue
		}

		reader.Close()
	}

	newFileName := filepath.Join(cm.dataDir, "sstable_"+time.Now().Format("20060102150405")+".db")
	builder, err := sstable.NewSSTableBuilder(newFileName)
	if err != nil {
		log.Printf("[Compaction] Failed to create compacted SSTable: %v", err)
		return
	}

	for k, v := range keyMap {
		if len(v) == 0 {
			continue 
		}
		builder.Add([]byte(k), v)
	}

	if err := builder.Finalize(); err != nil {
		log.Printf("[Compaction] Failed to finalize compacted SSTable: %v", err)
		return
	}

	for _, filePath := range sstFiles {
		os.Remove(filePath)
	}

	updateSSTables([]string{newFileName})
	log.Printf("[Compaction] Completed successfully. New file: %s", newFileName)
}