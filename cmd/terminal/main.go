package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rivo/tview"
	"greed"
	"greed/internal"
	"greed/pkg"
)

func main() {
	config := greed.Config()

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
	listener, err := pkg.NewListener()
	terminal := internal.NewTerminal(app, err == nil)
	terminal.BindMonitor(monitor)
	if err == nil {
		terminal.BindListener(listener)
		monitor.AddListener(listener)
	}

	go monitor.StartMonitoring()

	if err := app.SetRoot(terminal.Layout, true).Run(); err != nil {
		panic(err)
	}
}