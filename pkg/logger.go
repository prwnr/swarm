package pkg

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	//ERROR_LOG prefix
	ErrorLog = 1
	//DEBUG_LOG prefix
	DebugLog = 2
	//WARNING_LOG prefix
	WarningLog = 3
)

// LogError writes to swarm log line marked as error.
func LogError(text string) {
	Log(text, ErrorLog)
}

// LogDebug writes to swarm log line marked as debug.
func LogDebug(text string) {
	Log(text, DebugLog)
}


// LogWarning writes to swarm log line marked as warning.
func LogWarning(text string) {
	Log(text, WarningLog)
}

// Log text to file
func Log(text string, group int) {
	f, err := os.OpenFile("swarm.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	prefix := fmt.Sprintf("%s: ", groupString(group))
	logger := log.New(f, prefix, log.LstdFlags)
	logger.Println(strings.Trim(text, "\n"))
}

func groupString(group int) string {
	groups := map[int]string{
		1: "ERROR",
		2: "DEBUG",
		3: "WARNING",
	}

	if s, ok := groups[group]; ok != false {
		return s
	}

	return "UNKNOWN"
}