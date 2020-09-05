package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var name string = "Default"

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

	go endCheck(endChan, pc, addr)

	fmt.Print("Please type your name: \t")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	name = scanner.Text()

	clearConsole()
	go Listen(pc)

	pc.WriteTo([]byte(name+" has joined the chat..."), addr)
	fmt.Println("To exit the chat, type 'exit' or use ctrl-C")

	for {
		bytePW, _ := terminal.ReadPassword(int(syscall.Stdin))
		if string(bytePW) == "exit" {
			pc.WriteTo([]byte(name+" has left the chat..."), addr)
			exec.Command("stty echo").Run()
			os.Exit(0)
		}
		if string(bytePW) != "" {
			_, err2 := pc.WriteTo([]byte(name+": "+string(bytePW)), addr)
			if err2 != nil {
				panic(err2)
			}
		}
	}
}

func endCheck(endChan chan os.Signal, pc net.PacketConn, addr *net.UDPAddr) {
	select {
	case sig := <-endChan:
		pc.WriteTo([]byte(name+" has left the chat..."), addr)
		fmt.Printf("Leaving chat *%s*\n", sig)
		exec.Command("stty echo").Run()
		os.Exit(0)
	}
}

func Listen(pc net.PacketConn) {
	for {
		buf := make([]byte, 1024)
		n, _, err := pc.ReadFrom(buf)

		if err != nil {
			panic(err)
		}
		if n != 0 {
			//fmt.Println(strings.TrimSuffix(addr.String(), ":8000") + ": " + string(buf))
			fmt.Println(string(buf))
		}
	}
}

func clearConsole() {
	cm := exec.Command("clear") //Linux example, its tested
	cm.Stdout = os.Stdout
	cm.Run()
}
