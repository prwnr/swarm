package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rivo/tview"
	"stream-monitor/pkg"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:        "localhost:6379",
		Password:    "",
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

	if err := app.SetRoot(terminal.Flex, true).Run(); err != nil {
		panic(err)
	}
}
