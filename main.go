package main

import (
	"github.com/triapex/auth/cmd/auth"
	"time"
)

//func init() {
//	// Here all other sub commands should be registered to the rootCmd
//	rootCmd.AddCommand(srvCmd)
//	rootCmd.AddCommand(migrationRoot)
//}

func main() {
	time.Local = nil
	cmd.Execute()
}
