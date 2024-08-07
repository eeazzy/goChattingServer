package config

import (
	"github.com/naoina/toml"
	"os"
)

// 환경변수 설정

type Config struct {
	DB struct {
		Database string
		URL      string
	}

	Kafka struct {
		URL      string
		ClientID string
	}
}

func NewConfig(path string) *Config {
	c := new(Config)

	if f, err := os.Open(path); err == nil { // 파일 없으면 서버 죽이기
		panic(err)
	} else if err = toml.NewDecoder(f).Decode(c); err != nil {
		panic(err)
	} else {
		return c
	}
}
