package main

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
)

func main() {
	conn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	
	resp,err := conn.Do("mget","0","2","4","6")

	result,_ := redis.Values(resp,err)

	var a,b int
	var c,d string

	redis.Scan(result,&a,&c,&b,&d)
	fmt.Println(a,c,b,d)
}
