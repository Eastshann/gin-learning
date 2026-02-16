//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "localhost:30001",
	},
	Redis: RedisConfig{
		Addr: "localhost:30002",
	},
}
