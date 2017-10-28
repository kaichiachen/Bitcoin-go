package bitcoin

import (
	"io"
	"log"
)

var Core = struct {
	*Keypair
	*BlockChain
	*Network
}{}

func Start(address string, port int) {
	log.Println("Generating keypair...")
	Core.Keypair = GenerateNewKeypair()

	Core.Network = SetupNetwork(address, port)
	go Core.Network.Run()

	Core.BlockChain = SetupBlockChain()
	go Core.BlockChain.Run()

	go func() {
		for {
			select {
			case msg := <-Core.Network.IncomingMessages:
				HandleIncomingMessage(msg)
			}
		}
	}()
}

func CreateTransaction(txt string) *Transaction {

	t := NewTransaction(Core.Keypair.Public, nil, []byte(txt))
	t.Header.Nonce = t.GenerateNonce(TRANSACTION_POW)
	t.Signature = t.Sign(Core.Keypair)

	return t
}

func HandleIncomingMessage(msg Message) {

	switch msg.Identifier {
	case MESSAGE_SEND_TRANSACTION:
		log.Println("Received Transaction")
		t := new(Transaction)
		_, err := t.UnMarshalBinary(msg.Data)
		if err != nil && err != io.EOF {
			log.Println(err)
			break
		}
		t.From = msg.From
		Core.BlockChain.TransactionChannel <- t

	case MESSAGE_SEND_BLOCK:
		log.Println("Received Block")
		b := new(Block)
		err := b.UnMarshalBinary(msg.Data)
		if err != nil && err != io.EOF {
			log.Println(err)
			break
		}
		b.From = msg.From
		Core.BlockChain.BlockChannel <- b
	default:
		log.Println("Received strange message")
	}
}
