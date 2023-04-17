package main

import (
	"os"
	"strconv"
)

func GetPort() int {
	port := os.Getenv("PORT")
	if port == "" {
		return 0
	}

	i, err := strconv.Atoi(port)
	if err != nil {
		return 0
	}

	return i
}
