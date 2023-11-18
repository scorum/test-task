package config

import "time"

type ServerConfig struct {
	Addr         string        `yaml:"addr" required:"true"`
	ReadTimeout  time.Duration `yaml:"http_read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"http_write_timeout" default:"30s"`
}

type DBConfig struct {
	Addr         string `yaml:"addr" required:"true"`
	MaxIdleConns int    `yaml:"max_idle_conns" default:"10"`
	MaxOpenConns int    `yaml:"max_open_conns" default:"10"`
}
