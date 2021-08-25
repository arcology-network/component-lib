package log

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Console struct {
	SystemOut bool
}

type LocalFile struct {
	SaveFile   bool
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Ignored    string
}

var Conf *Config

// Config .
type Config struct {
	Level     string
	Console   *Console
	LocalFile *LocalFile
}

func InitCfg(confPath string) {
	_, err := toml.DecodeFile(confPath, &Conf)
	if err != nil {
		fmt.Printf("err=%v\n", err)
	}
	fmt.Printf("conf=%v,%v\n", Conf.Console, Conf.LocalFile)
}
