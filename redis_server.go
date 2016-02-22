//package the redis client with Golang
package main

import (
	"fmt"
	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
	"math/rand"
	"time"
)

//type ClientPool struct {
//	redispool *pool.Pool
//}

//type RedisClient struct{
//}
//
//func ReleaseConnection() {
//	""
//}

type MMpool struct {
	*pool.Pool
}

func (p *MMpool) GetConnection() *redis.Client {
	client, err := p.Pool.Get()
	if err != nil {
		// handle error
	}
	return client
}

func ConnectionPool() *MMpool {
	p, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		// handle error
	}
	return &MMpool{p}
}

func AddPlayer(c *redis.Client, playerlevel int64, playerid string) {
	v, err := c.Cmd("ZADD", "battle_level", playerlevel, playerid).Str()
	if err != nil {
		//handle error
	}
	fmt.Println(v)
}

func GetPlayer(c *redis.Client, playerlevel int64, vsrange int64) string {
	range_floor := playerlevel - vsrange
	if range_floor < 0 {
		range_floor = 0
	}
	v, err := c.Cmd("ZREVRANGEBYSCORE", "battle_level", playerlevel+vsrange, range_floor).List()
	if err != nil {
		//Handle the error
	}
	size := len(v)
	rand.Seed(time.Now().Unix())
	arrayindex := rand.Intn(size)
	fmt.Println(v)
	matchedPlayer := v[arrayindex]
	defer remPlayer(c, v[arrayindex])
	return matchedPlayer
}

func remPlayer(c *redis.Client, playerid string) {
	v, err := c.Cmd("ZREM", "battle_level", playerid).Str()
	if err != nil {
		//Handle the error
	}
	fmt.Println(v)
}
