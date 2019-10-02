package stream_monitor

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// Config of the Monitor
func Config() Configuration {
	jsonFile, err := os.Open("config.json")
	config := Configuration{
		RedisHost: "localhost",
		RedisPort: 6379,
	}

	if err == nil {
		defer jsonFile.Close()
		bytes, _ := ioutil.ReadAll(jsonFile)
		_ = json.Unmarshal(bytes, &config)
	}

	return config
}

// Configuration values
type Configuration struct {
	RedisHost     string `json:"redis_host,omitempty"`
	RedisPort     int    `json:"redis_port,omitempty"`
	RedisPassword string `json:"redis_password,omitempty"`
	ArtisanPath   string `json:"artisan_path,omitempty"`
}
