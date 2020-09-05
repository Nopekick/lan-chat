package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	cmd := "ifconfig | grep netmask"
	out, err := exec.Command("bash", "-c", cmd).Output()

	if err != nil {
		panic("command failed :(")
	}

	re := regexp.MustCompile(`broadcast (.+)`)
	found := re.FindSubmatch([]byte(string(out)))
	broadcastAddr := string(found[1]) + ":8000"

	pc, err := net.ListenPacket("udp4", ":8000")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	addr, err := net.ResolveUDPAddr("udp4", broadcastAddr)
	if err != nil {
		panic(err)
	}

	go Listen(pc)

	for {
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		if string(bytePassword) != "" {
			_, err2 := pc.WriteTo([]byte(string(bytePassword)), addr)
			if err2 != nil {
				panic(err2)
			}
		}
	}
}

func Listen(pc net.PacketConn) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := pc.ReadFrom(buf)

		if err != nil {
			panic(err)
		}
		if n != 0 {
			fmt.Println(strings.TrimSuffix(addr.String(), ":8000") + ": " + string(buf))
		}
	}
}
