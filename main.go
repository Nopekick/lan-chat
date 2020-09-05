package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
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

	endChan := make(chan os.Signal)
	signal.Notify(endChan, os.Interrupt)

	go endCheck(endChan, pc)
	go Listen(pc)

	for {
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		if string(bytePassword) == "exit" {
			exec.Command("stty echo").Run()
			os.Exit(0)
		}
		if string(bytePassword) != "" {
			_, err2 := pc.WriteTo([]byte(string(bytePassword)), addr)
			if err2 != nil {
				panic(err2)
			}
		}
	}
}

func endCheck(endChan chan os.Signal, pc net.PacketConn) {
	select {
	case sig := <-endChan:
		exec.Command("stty echo").Run()
		pc.Close()
		fmt.Printf("<<%s>> Leaving chat...\n", sig)
		os.Exit(0)
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
