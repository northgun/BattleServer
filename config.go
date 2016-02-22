package main

type Config struct {
	Port          string
	Certfile      string
	Keyfile       string
	CGSEndpoint   string
	LoginEndpoint string
	RedisAddr     string
	RedisPassword string
	RedisDatabase int64
}
