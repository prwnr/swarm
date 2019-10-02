package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rivo/tview"
	stream_monitor "stream-monitor"
	"stream-monitor/pkg"
)

func main() {
	config := stream_monitor.Config()

	client := redis.NewClient(&redis.Options{
		Addr:        fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password:    config.RedisPassword,
		DB:          0,
		ReadTimeout: -1,
	})

	_, err := client.Ping().Result()
	if err != nil {
		panic(fmt.Sprintf("failed to connect with Redis, err: %v", err))
	}

	app := tview.NewApplication()
	monitor := pkg.NewMonitor(client)
	terminal := pkg.NewTerminal(app)
	terminal.BindMonitor(monitor)

	go monitor.StartMonitoring()

	if err := app.SetRoot(terminal.Layout, true).Run(); err != nil {
		panic(err)
	}
}