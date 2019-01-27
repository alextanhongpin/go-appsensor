package main

import (
	"fmt"
	"strings"
)

func NewID(clientIP string, rest ...string) string {
	return strings.Join(append([]string{clientIP}, rest...), "/")
}

func main() {
	// Prioritize on the ClientIP, since the UserID might not be available
	// if the endpoint is a public ip. 
	// If we are using sessions, we can always use the SessionID to block the user.
	fmt.Println(NewID("0.0.0.0", "user-id-1", "appName"))
	fmt.Println(NewID("0.0.0.0", "", "appName"))
	fmt.Println(NewID("0.0.0.0"))
}
