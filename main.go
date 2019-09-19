package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"strings"
)

func main()  {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		ReadTimeout: -1,
	})

	var events []string
	var errors int

	for {
		res, err := client.Do("MONITOR").String()
		if err != nil {
			fmt.Errorf(err.Error())
			errors++
			if errors > 5 {
				panic(fmt.Sprintf("MONITOR keeps failing, last error: %v", err))
			}

			continue
		}

		split := strings.Split(strings.Replace(res, "\"", "", -1), " ")
		//fmt.Printf("%v \r\n", split)
		if len(split) > 3 && strings.ToUpper(split[3]) == "XADD" {
			e := split[4]
			if sliceContains(events, e) {
				continue
			}

			events = append(events, e)
			go startListening(client, e)
		}
	}
}

func startListening(client *redis.Client, stream string) {
	fmt.Printf("%v \r\n", stream)
	messages, _ := client.XRange(stream, "-", "+").Result()
	for _, m := range messages {
		fmt.Printf("Stream: %s\r\n", stream)
		fmt.Printf("ID: %s: %v\r\n", m.ID, m.Values)
	}

	for {
		newMessages, _ := client.XRead(&redis.XReadArgs{
			Streams: []string{stream, "$"},
			Block:   0,
		}).Result()
		for _, xStream := range newMessages {
			fmt.Printf("Stream: %s\r\n", xStream.Stream)
			for _, m := range xStream.Messages {
				fmt.Printf("ID: %s: %v\r\n", m.ID, m.Values)
			}
		}
	}
}

// SliceContains checks if slice of strings contains given string
func sliceContains(s []string, l string) bool {
	for _, v := range s {
		if v == l {
			return true
		}
	}

	return false
}