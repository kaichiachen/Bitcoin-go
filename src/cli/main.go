package main

import (
	"bitcoin"
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"regexp"
)

var port int
var slow bool

func init() {
	flag.IntVar(&port, "port", bitcoin.BLOCKCHAIN_DEFAULT_PORT, "blockchain port")
	flag.BoolVar(&slow, "slow", false, "POW speed")
}

func usage() {
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	bitcoin.Start(getIPAddress(), port)
	for {
		input := <-readStdin()
		if addr := findIPAddress(input); addr != "" {
			bitcoin.Core.Network.ConnectionQueue <- addr
		} else {
			bitcoin.Core.BlockChain.TransactionChannel <- bitcoin.CreateTransaction(input)
		}
	}
}

func readStdin() chan string {
	cb := make(chan string)
	input := bufio.NewScanner(os.Stdin)
	go func() {
		if input.Scan() {
			cb <- input.Text()
		}
	}()

	return cb
}

func getIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func findIPAddress(input string) string {
	validIpAddressRegex := "([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})"
	re := regexp.MustCompile(validIpAddressRegex)
	return re.FindString(input)
}
