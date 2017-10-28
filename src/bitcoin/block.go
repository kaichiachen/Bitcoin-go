package bitcoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"reflect"
)

type BlockHeader struct {
	Origin     []byte
	PrevBlock  []byte
	MerkelRoot []byte
	Timestamp  uint32
	Nonce      uint32
}

func (bh *BlockHeader) MarshalBinary() ([]byte, error) {
	bs := &bytes.Buffer{}

	bs.Write(FitBytes(bh.Origin, NETWORK_KEY_SIZE))
	bs.Write(FitBytes(bh.PrevBlock, 32))
	bs.Write(FitBytes(bh.MerkelRoot, 32))
	binary.Write(bs, binary.LittleEndian, bh.Timestamp)
	binary.Write(bs, binary.LittleEndian, bh.Nonce)

	return bs.Bytes(), nil
}

func (bh *BlockHeader) UnMarshalBinary(d []byte) error {
	bf := bytes.NewBuffer(d)

	bh.Origin = bf.Next(NETWORK_KEY_SIZE)
	bh.PrevBlock = bf.Next(32)
	bh.PrevBlock = bf.Next(32)
	binary.Read(bytes.NewBuffer(bf.Next(4)), binary.LittleEndian, &bh.Timestamp)
	binary.Read(bytes.NewBuffer(bf.Next(4)), binary.LittleEndian, &bh.Nonce)

	return nil
}

type BlockSlice []Block

func (bs BlockSlice) Exists(b Block) bool {

	l := len(bs)
	for i := l - 1; i >= 0; i-- {

		bb := bs[i]
		if reflect.DeepEqual(b.Signature, bb.Signature) {
			return true
		}
	}
	return false
}

func (bs BlockSlice) PreviousBlock() *Block {
	l := len(bs)
	if l == 0 {
		return nil
	} else {
		return &bs[l-1]
	}
}

type Block struct {
	*BlockHeader
	Signature []byte
	*TransactionSlice
	From []byte
}

func NewBlock(previousBlock []byte) Block {
	header := &BlockHeader{PrevBlock: previousBlock}
	return Block{header, nil, new(TransactionSlice), nil}
}

func (b *Block) AddTransaction(tr Transaction) {
	slice := b.TransactionSlice.AddTransaction(tr)
	b.TransactionSlice = &slice
}

func (b *Block) Hash() []byte {
	headerHash, _ := b.BlockHeader.MarshalBinary()
	hash := sha256.New()
	hash.Write(headerHash)
	return hash.Sum(nil)
}

func (b *Block) Sign(keypair *Keypair) []byte {
	s, _ := keypair.Sign(b.Hash())
	return s
}

func (b *Block) VerifyBlock(prefix []byte) bool {
	headerHash := b.Hash()
	merkel := b.GenerateMerkelRoot()

	if !reflect.DeepEqual(merkel, b.BlockHeader.MerkelRoot) {
		log.Println("merkel wrong")
		return false
	}

	if !CheckProofOfWork(prefix, headerHash) {
		log.Println("Fail to check proof of work")
		return false
	}

	if !SignatureVerify(b.BlockHeader.Origin, b.Signature, headerHash) {
		log.Println("Signature verfication fail")
		return false
	}
	return true
}

func (b *Block) GenerateMerkelRoot() []byte {
	var merkell func(hashes [][]byte) []byte
	merkell = func(hashes [][]byte) []byte {
		l := len(hashes)
		if l == 0 {
			return nil
		} else if l == 1 {
			return hashes[0]
		} else {
			if l%2 == 1 {
				return merkell([][]byte{merkell(hashes[:l-1]), hashes[l-1]})
			}

			bs := make([][]byte, l/2)
			for i, _ := range bs {
				j, k := i*2, (i*2)+1
				hash := sha256.New()
				hash.Write(append(hashes[j], hashes[k]...))
				bs[i] = hash.Sum(nil)
			}
			return merkell(bs)
		}
	}

	var ts [][]byte
	for _, v := range *b.TransactionSlice {
		ts = append(ts, v.Hash())
	}

	return merkell(ts)
}

func (b Block) GenerateNounce(prefix []byte) uint32 {

	for {
		if CheckProofOfWork(prefix, b.Hash()) {
			break
		}
		b.BlockHeader.Nonce++
	}
	return b.BlockHeader.Nonce
}

func (b *Block) MarshalBinary() ([]byte, error) {

	bs := bytes.Buffer{}

	bhb, err := b.BlockHeader.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bs.Write(bhb)
	bs.Write(FitBytes(b.Signature, NETWORK_KEY_SIZE))

	tsb, err := b.TransactionSlice.MarshalBinary()
	if err != nil {
		return nil, err
	}
	bs.Write(FitBytes(b.From, IP_SIZE))
	bs.Write(tsb)

	return bs.Bytes(), nil
}

func (b *Block) UnMarshalBinary(d []byte) error {

	buf := bytes.NewBuffer(d)

	header := new(BlockHeader)
	err := header.UnMarshalBinary(buf.Next(BLOCK_HEADER_SIZE))
	if err != nil {
		return err
	}

	b.BlockHeader = header
	b.Signature = buf.Next(NETWORK_KEY_SIZE)
	b.From = buf.Next(IP_SIZE)

	ts := new(TransactionSlice)
	err = ts.UnMarshalBinary(buf.Next(math.MaxInt64))
	if err != nil {
		return err
	}

	b.TransactionSlice = ts

	return nil
}
