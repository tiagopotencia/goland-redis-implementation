package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"gopkg.in/redsync.v1"

	"github.com/garyburd/redigo/redis"
)

///ATK to be inserted
type AtkToBeInserted struct {
	Atk string `json: "Atk"`
}

var (
	Pool redsync.Pool
	sync *redsync.Redsync
)

func init() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = ":6379"
	}
	Pool = newPool(redisHost)
}

func newPool(server string) redsync.Pool {

	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func main() {

	initEndpoints()

	var pools []redsync.Pool

	pool2 := newPool(":6379")

	pools = append(pools, Pool)
	pools = append(pools, pool2)
	sync = redsync.New(pools)

}

func insert(atk string) {

	conn := Pool.Get()
	defer conn.Close()

	fmt.Println("Try to insert atk " + atk)

	data, err := redis.String(conn.Do("SET", "atk", atk, "NX"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("%v", data))

	data, err = redis.String(conn.Do("GET", "atk"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("%v", data))
}

func lock() {

	mutex := sync.NewMutex("atk")
	err := mutex.Lock()

	if err != nil {
		fmt.Println(fmt.Sprintf("%+v", err))
	} else {
		fmt.Println("Locked!")
	}

}

func initEndpoints() {

	engine := gin.New()

	engine.POST("/", postAtk)

	engine.Run(":5473")
}

func postAtk(c *gin.Context) {

	var atk AtkToBeInserted

	err := c.Bind(&atk)

	if err != nil {
		fmt.Println(err)
	}

	insert(atk.Atk)

}
