//go:build ignore
// +build ignore

package main

import (
	"log"
	"time"

	"xelf.org/layla/tsc"
)

func main() {
	c, err := tsc.Auto("/dev/usb/lp1", time.Second)
	if err != nil {
		log.Fatal(err)
	}
	stat, err := c.Status()
	if err != nil {
		log.Fatal(err)
	}
	name, err := c.Model()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Printer %s %s!\n", name, stat)
}
