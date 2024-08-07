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

	if f, err := os.Open(path); err != nil { // 파일 열기에 실패하면 패닉
		panic(err)
	} else {
		defer f.Close() // 파일 닫기
		if err = toml.NewDecoder(f).Decode(c); err != nil {
			panic(err)
		}
		return c
	}
}
