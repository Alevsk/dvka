package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		message := os.Args[1]
		fmt.Print(base64.StdEncoding.EncodeToString([]byte(message)))
	}
}