package typewriter

import "os"

type Config struct {
	Filter                func(os.FileInfo) bool
}

var DefaultConfig = &Config{}
