package bitcoin

import (
	"log"
	"reflect"
	"time"
)

type TransactionChannel chan *Transaction
type BlockChannel chan *Block
type BlockChain struct {
	CurrentBlock Block
	BlockSlice

	TransactionChannel
	BlockChannel
}

func SetupBlockChain() *BlockChain {

	bl := new(BlockChain)
	bl.TransactionChannel, bl.BlockChannel = make(TransactionChannel), make(BlockChannel)

	//Read blockchain from file and stuff...

	bl.CurrentBlock = bl.CreateNewBlock()

	return bl
}

func (bl *BlockChain) CreateNewBlock() Block {
	prevBlock := bl.BlockSlice.PreviousBlock()
	prevBlockHash := []byte{}

	if prevBlock != nil {
		prevBlockHash = prevBlock.Hash()
	}

	b := NewBlock(prevBlockHash)
	b.BlockHeader.Origin = Core.Keypair.Public

	return b
}

func (bl *BlockChain) AddBlock(b Block) {
	bl.BlockSlice = append(bl.BlockSlice, b)
}

func (bl *BlockChain) Run() {

	interruptBlockGen := bl.GenerateBlocks()

	for {
		select {
		case tr := <-bl.TransactionChannel:
			if bl.CurrentBlock.TransactionSlice.Exists(*tr) {
				continue
			}
			if !tr.VerifyTransaction(TRANSACTION_POW) {
				log.Println("Received non valid transaction", tr)
				continue
			}

			bl.CurrentBlock.AddTransaction(*tr)
			interruptBlockGen <- bl.CurrentBlock

			mes := NewMessage(MESSAGE_SEND_TRANSACTION)
			mes.Data, _ = tr.MarshalBinary()
			mes.From = tr.From

			time.Sleep(300 * time.Millisecond)
			Core.Network.BroadcastQueue <- *mes

		case b := <-bl.BlockChannel:

			if bl.BlockSlice.Exists(*b) {
				log.Println("block exists")
				continue
			}

			if !b.VerifyBlock(BLOCK_POW) {
				log.Println("block verification fails")
				continue
			}

			if reflect.DeepEqual(b.PrevBlock, bl.CurrentBlock.Hash()) {
				log.Println("Missing blocks in between")
			} else {

				log.Println("New block!", b.Hash())

				transDiff := TransactionSlice{}
				if !reflect.DeepEqual(b.BlockHeader.MerkelRoot, bl.CurrentBlock.MerkelRoot) {
					// Transactions are different
					log.Println("Transactions are different. finding diff")
					transDiff = DiffTransactionSlices(*bl.CurrentBlock.TransactionSlice, *b.TransactionSlice)
				}

				bl.AddBlock(*b)

				mes := NewMessage(MESSAGE_SEND_BLOCK)
				mes.Data, _ = b.MarshalBinary()
				mes.From = b.From
				Core.Network.BroadcastQueue <- *mes

				bl.CurrentBlock = bl.CreateNewBlock()
				bl.CurrentBlock.TransactionSlice = &transDiff

				interruptBlockGen <- bl.CurrentBlock
			}
		}
	}
}

func DiffTransactionSlices(a, b TransactionSlice) (diff TransactionSlice) {
	//Assumes transaction arrays are sorted (which maybe is too big of an assumption)
	lastj := 0
	for _, t := range a {
		found := false
		for j := lastj; j < len(b); j++ {
			if reflect.DeepEqual(b[j].Signature, t.Signature) {
				found = true
				lastj = j
				break
			}
		}
		if !found {
			diff = append(diff, t)
		}
	}

	return
}

func (bl *BlockChain) GenerateBlocks() chan Block {
	interrupt := make(chan Block)

	go func() {
		block := <-interrupt
	loop:
		log.Println("Starting Proof of Work...")
		block.BlockHeader.MerkelRoot = block.GenerateMerkelRoot()
		block.BlockHeader.Nonce = 0
		block.BlockHeader.Timestamp = uint32(time.Now().Unix())

		for true {
			sleepTime := time.Nanosecond
			if block.TransactionSlice.Len() > 0 {
				if CheckProofOfWork(BLOCK_POW, block.Hash()) {
					block.Signature = block.Sign(Core.Keypair)
					bl.BlockChannel <- &block
					sleepTime = time.Hour * 24
					log.Println("Found Block!!")
				} else {
					block.BlockHeader.Nonce += 1
				}
			} else {
				sleepTime = time.Hour * 24
				log.Println("No trans sleep")
			}

			select {
			case block = <-interrupt:
				goto loop
			case <-time.NewTimer(sleepTime).C:
				continue
			}

		}
	}()
	return interrupt
}
