//go:build ignore

package main

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

func main() {
	var c *whatsmeow.Client
	var j types.JID
	// Testing for a potential GetPhoneForLID method on the client
	_ = c.GetPhoneForLID(j)
}
