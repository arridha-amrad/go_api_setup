package main

import (
	"fmt"
	"time"
)

func main() {
	due := time.Now().Add(1 * time.Hour).UnixMilli()
	fmt.Println(due)
}
