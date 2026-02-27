package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/valkey-io/valkey-go"
)

func main() {
	f, _ := os.Open(".env")
	scanner := bufio.NewScanner(f)
	var url string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VALKEY_URL=") {
			url = strings.TrimPrefix(line, "VALKEY_URL=")
			break
		}
	}
	f.Close()

	fmt.Printf("Testing URL: %s\n", url)

	opts, err := valkey.ParseURL(url)
	if err != nil {
		fmt.Printf("ParseURL Error: %v\n", err)
		return
	}

	fmt.Printf("Parsed Host: %v\n", opts.InitAddress)
	fmt.Printf("Parsed Password: %s\n", opts.Password)

	client, err := valkey.NewClient(opts)
	if err != nil {
		fmt.Printf("NewClient Error: %v\n", err)
		return
	}
	defer client.Close()

	err = client.Do(context.Background(), client.B().Ping().Build()).Error()
	if err != nil {
		fmt.Printf("Do Error: %v\n", err)
	} else {
		fmt.Println("Success!")
	}
}
