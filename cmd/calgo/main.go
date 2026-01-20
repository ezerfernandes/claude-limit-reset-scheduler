package main

import (
	"fmt"
	"os"

	// Dependencies - will be used in subsequent tasks
	_ "github.com/araddon/dateparse"
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/viper"
	_ "golang.org/x/oauth2/google"
	_ "google.golang.org/api/calendar/v3"
)

var version = "0.1.0"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("calgo version %s\n", version)
		return
	}

	fmt.Println("calgo - Google Calendar CLI tool")
	fmt.Println("Usage: calgo [command] [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create    Create a new calendar event")
	fmt.Println("  quick     Create event using natural language")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --help, -h       Show help")
	fmt.Println("  --version, -v    Show version")
}
