package db

type Options struct {
	DataDir         string
	MaxMemtableSize int64 
}

func DefaultOptions(dataDir string) Options {
	return Options{
		DataDir:         dataDir,
		MaxMemtableSize: 4 * 1024 * 1024, 
	}
}