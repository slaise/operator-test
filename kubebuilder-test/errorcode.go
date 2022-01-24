package main

import "fmt"

type ErrorCode int

//go:generate stringer -type ErrorCode -linecomment
const (
	OK        ErrorCode = iota // OK
	NOT_FOUND                  // Resource Not Found
	TIMEOUT                    // Call Timeout
)

func main() {
	fmt.Println(NOT_FOUND)
}
