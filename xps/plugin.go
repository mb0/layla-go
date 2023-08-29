package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "xelf.org/layla"
	_ "xelf.org/layla/font"
	_ "xelf.org/layla/html"
	_ "xelf.org/layla/pdf"
	"xelf.org/layla/tsc"
	_ "xelf.org/layla/tsc"
	_ "xelf.org/layla/tspl"
	"xelf.org/xelf/xps"
	_ "xelf.org/xelf/xps"
)

func Cmd(ctx *xps.CmdCtx) error {
	switch ctx.Split() {
	case "tscscan":
		devs, err := tsc.DiscoverDev()
		if err != nil {
			log.Printf("error scanning for local printers: %v", err)
		} else if len(devs) > 0 {
			fmt.Printf("# local printers:\n")
			for _, dev := range devs {
				fmt.Printf("%s\n", dev)
			}
		} else {
			fmt.Printf("# no local printers\n")
		}
		nfos, err := tsc.DiscoverNet(1 * time.Second)
		if err != nil {
			log.Printf("error scanning for network printers: %v", err)
		} else if len(nfos) > 0 {
			fmt.Printf("# network printers:\n")
			for _, nfo := range nfos {
				fmt.Printf("%s\n", nfo.IP)
			}
		} else {
			fmt.Printf("# no network printers\n")
		}
	case "raw":
		return printRaw(ctx.Args)
	}
	return nil
}

func decImgFile(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	return image.Decode(f)
}

func printRaw(args []string) error {
	var r io.ReadCloser
	if n := len(args); n == 0 || n == 1 && args[0] == "--" {
		r = os.Stdin
	} else if n == 1 {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		r = f
	} else {
		return fmt.Errorf("expect raw file or stdin")
	}
	defer r.Close()
	c, err := tsc.Auto("", 0)
	if err != nil {
		return err
	}
	defer c.Close()
	_, err = io.Copy(c.Conn, r)
	return nil
}
