package bitcoin

import (
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"time"
)

type ConnectionQueue chan string
type NodeChannel chan *Node
type Node struct {
	*net.TCPConn
	lastSeen int
}

type Nodes map[string]*Node
type Network struct {
	Nodes
	ConnectionQueue
	Address            string
	ConnectionCallback NodeChannel
	BroadcastQueue     chan Message
	IncomingMessages   chan Message
}

func SetupNetwork(address string, port int) *Network {

	n := &Network{}

	n.BroadcastQueue, n.IncomingMessages = make(chan Message), make(chan Message)
	n.ConnectionQueue, n.ConnectionCallback = CreateConnectionQueue(port)
	n.Nodes = Nodes{}
	n.Address = fmt.Sprintf("%s:%d", address, port)

	return n
}

func CreateConnectionQueue(port int) (ConnectionQueue, NodeChannel) {
	in := make(ConnectionQueue)
	out := make(NodeChannel)

	go func() {

		for {
			address := <-in
			address = fmt.Sprintf("%s:%d", address, port)
			log.Println(address)
			if address != Core.Network.Address && Core.Nodes[address] == nil {
				log.Printf("Connect to node: %s\n", address)
				go ConnectToNode(address, 5*time.Second, false, out)
			}
		}
	}()

	return in, out
}

func (n *Network) Run() {

	log.Println("Listening in", Core.Network.Address)
	listenCb := StartListening(Core.Network.Address)

	for {
		select {
		case node := <-listenCb:
			Core.Nodes.AddNode(node)

		case node := <-n.ConnectionCallback:
			Core.Nodes.AddNode(node)

		case message := <-n.BroadcastQueue:
			go n.BroadcastMessage(message)
		}
	}
}

func (n Nodes) AddNode(node *Node) bool {
	ip, _, _ := net.SplitHostPort(node.TCPConn.RemoteAddr().String())
	addr := fmt.Sprintf("%s:%d", ip, BLOCKCHAIN_DEFAULT_PORT)

	if addr != Core.Network.Address && n[addr] == nil {
		log.Println("Node connected ", addr)
		n[addr] = node

		go HandleNode(node)

		return true
	}

	log.Println("Duplicate ip address")
	return false
}

func StartListening(address string) NodeChannel {

	cb := make(NodeChannel)
	addr, err := net.ResolveTCPAddr("tcp4", address)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}

	go func(l *net.TCPListener) {

		for {
			connection, err := l.AcceptTCP()
			if err != nil && err != io.EOF {
				log.Println(err)
			}

			cb <- &Node{connection, int(time.Now().Unix())}
		}
	}(listener)

	return cb
}

func HandleNode(node *Node) {
	for {
		var bs []byte = make([]byte, 1024*1000)
		n, err := node.TCPConn.Read(bs)
		if err == io.EOF {
			log.Printf("%s： Connection Closed\n", node.TCPConn.RemoteAddr().String())
			node.TCPConn.Close()
			break
		}
		if err != nil {
			log.Println("Blockchain network: ", err)
		}

		m := new(Message)
		err = m.UnMarshalBinary(bs[:n])

		if err != nil {
			log.Println(err)
			continue
		}

		m.Reply = make(chan Message)

		go func(cb chan Message) {
			for {
				m, ok := <-cb

				if !ok {
					close(cb)
					break
				}

				b, _ := m.MarshalBinary()
				l := len(b)

				i := 0
				for i < l {
					a, _ := node.TCPConn.Write(b[i:])
					i += a
				}
			}
		}(m.Reply)

		Core.Network.IncomingMessages <- *m
	}
}

func ConnectToNode(dst string, timeout time.Duration, retry bool, cb NodeChannel) {

	addrDst, err := net.ResolveTCPAddr("tcp4", dst)

	if err != nil && err != io.EOF {
		log.Println(err)
	}

	var con *net.TCPConn = nil
loop:
	for {
		breakChannel := make(chan bool)
		go func() {

			con, err = net.DialTCP("tcp", nil, addrDst)

			if con != nil {
				cb <- &Node{con, int(time.Now().Unix())}
				breakChannel <- true
			}
		}()

		select {
		case <-time.NewTimer(timeout).C:
			if !retry {
				break loop
			}
		case <-breakChannel:
			break loop
		}
	}
}

func (n *Network) BroadcastMessage(message Message) {
	originalFrom := message.From
	message.From = []byte(n.Address)
	b, _ := message.MarshalBinary()

	for k, node := range n.Nodes {
		ip, _, _ := net.SplitHostPort(node.TCPConn.RemoteAddr().String())
		nodeAddr := fmt.Sprintf("%s:%d", ip, BLOCKCHAIN_DEFAULT_PORT)

		if reflect.DeepEqual(originalFrom, FitBytes([]byte(nodeAddr), IP_SIZE)) {
			continue
		}
		log.Println("Broadcasting...", k)
		go func() {
			_, err := node.TCPConn.Write(b)
			if err != nil {
				log.Println("Error bcing to", node.TCPConn.RemoteAddr())
			}
		}()
	}
}
