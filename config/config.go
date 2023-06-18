package config

import (
	"github.com/uerax/goconf"
)

func Load(path string) {
	goconf.LoadConfig(path)
}