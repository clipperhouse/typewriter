package typewriter

import "os"

type Config struct {
	Filter                func(os.FileInfo) bool
	IgnoreTypeCheckErrors bool
	OneFile               bool
}

var DefaultConfig = &Config{}
