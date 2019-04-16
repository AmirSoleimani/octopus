package main

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		// Addr:     "172.20.4.101:9090",
		Addr:     "localhost:9090",
		Password: "", // no password set
		DB:       0,  // use default DB
		// ReadTimeout: 5 * time.Second,
		// PoolSize:    40,
	})

	time.Sleep(2 * time.Second)
RETRY:
	count := 0
	for i := 1; i <= 100; i++ {
		if i%70 == 0 {
			time.Sleep(1 * time.Second)
		}
		go func(i int) {
			start := time.Now().Unix()
			fmt.Println(client.Ping())
			end := time.Now().Unix()
			diff := end - start
			if diff >= 1 {
				fmt.Println("========")
				fmt.Println("DIFF", diff)
				fmt.Println("========")
			}
			// fmt.Println(client.SMembers("peste:category:6").Result())
		}(i)
	}

	fmt.Println(count)
	goto RETRY
}
