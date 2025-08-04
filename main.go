package main

import (
	cmd "github.com/fahimimam/invoice/cmd/invoice"
	"time"
)

func main() {
	time.Local = nil
	cmd.Execute()
}
