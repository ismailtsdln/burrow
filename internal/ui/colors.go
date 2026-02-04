package ui

import "fmt"

const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

// Colorize returns the string wrapped in the given color code.
func Colorize(color, text string) string {
	return color + text + Reset
}

// PrintSuccess prints a message in green with a checkmark.
func PrintSuccess(format string, a ...interface{}) {
	fmt.Printf(Green+"✅ "+format+Reset+"\n", a...)
}

// PrintError prints a message in red with an X mark.
func PrintError(format string, a ...interface{}) {
	fmt.Printf(Red+"❌ "+format+Reset+"\n", a...)
}

// PrintWarning prints a message in yellow with a warning sign.
func PrintWarning(format string, a ...interface{}) {
	fmt.Printf(Yellow+"⚠️  "+format+Reset+"\n", a...)
}

// PrintInfo prints a message in blue/cyan with an info icon.
func PrintInfo(format string, a ...interface{}) {
	fmt.Printf(Cyan+"ℹ️  "+format+Reset+"\n", a...)
}

// PrintHeader prints a bold header.
func PrintHeader(title string) {
	fmt.Println("\n" + Bold + title + Reset)
}
