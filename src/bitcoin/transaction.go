package bitcoin

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"time"
)

type Transaction struct {
	Header    TransactionHeader
	Signature []byte
	Payload   []byte
	From      []byte
}

type TransactionHeader struct {
	From          []byte
	To            []byte
	Timestamp     uint32
	Nonce         uint32
	PayloadHash   []byte
	PayloadLength uint32
}

func NewTransaction(from, to, payload []byte) *Transaction {
	t := Transaction{
		Header: TransactionHeader{
			From: from,
			To:   to,
		},
		Payload: payload,
	}

	t.Header.Timestamp = uint32(time.Now().Unix())
	hash := sha256.New()
	hash.Write(payload)
	t.Header.PayloadHash = hash.Sum(nil)
	t.Header.PayloadLength = uint32(len(t.Payload))

	return &t
}

func (t *Transaction) Hash() []byte {
	headerBytes, _ := t.Header.MarshalBinary()
	hash := sha256.New()
	hash.Write(headerBytes)

	return hash.Sum(nil)
}

func (t *Transaction) Sign(keypair *Keypair) []byte {

	s, _ := keypair.Sign(t.Hash())
	return s
}

func (t *Transaction) VerifyTransaction(pow []byte) bool {

	headerHash := t.Hash()
	hash := sha256.New()
	hash.Write(t.Payload)
	payloadHash := hash.Sum(nil)

	return reflect.DeepEqual(payloadHash, t.Header.PayloadHash) &&
		CheckProofOfWork(pow, headerHash) &&
		SignatureVerify(t.Header.From, t.Signature, headerHash)

}

func (t Transaction) GenerateNonce(prefix []byte) uint32 {
	for {
		if CheckProofOfWork(prefix, t.Hash()) {
			break
		}
		t.Header.Nonce++
	}
	return t.Header.Nonce
}

func (th *TransactionHeader) MarshalBinary() ([]byte, error) {
	buf := &bytes.Buffer{}

	buf.Write(FitBytes(th.From, NETWORK_KEY_SIZE))
	buf.Write(FitBytes(th.To, NETWORK_KEY_SIZE))
	binary.Write(buf, binary.LittleEndian, th.Timestamp)
	buf.Write(th.PayloadHash[:32])
	binary.Write(buf, binary.LittleEndian, th.PayloadLength)
	binary.Write(buf, binary.LittleEndian, th.Nonce)

	return buf.Bytes(), nil
}

func (th *TransactionHeader) UnMarshalBinary(d []byte) error {
	buf := bytes.NewBuffer(d)
	th.From = buf.Next(NETWORK_KEY_SIZE)
	th.To = buf.Next(NETWORK_KEY_SIZE)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Timestamp)
	th.PayloadHash = buf.Next(32)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.PayloadLength)
	binary.Read(bytes.NewBuffer(buf.Next(4)), binary.LittleEndian, &th.Nonce)

	return nil
}

func (t *Transaction) MarshalBinary() ([]byte, error) {

	bs := &bytes.Buffer{}

	thBytes, _ := t.Header.MarshalBinary()
	if len(thBytes) != TRANSACTION_HEADER_SIZE {
		return nil, errors.New("Wrong Byte length")
	}
	bs.Write(thBytes)
	bs.Write(FitBytes(t.Signature, NETWORK_KEY_SIZE))
	bs.Write(t.Payload)
	bs.Write(FitBytes(t.From, IP_SIZE))

	return bs.Bytes(), nil
}

func (t *Transaction) UnMarshalBinary(d []byte) ([]byte, error) {

	bf := bytes.NewBuffer(d)
	thBytes := bf.Next(TRANSACTION_HEADER_SIZE)
	t.Header.UnMarshalBinary(thBytes)
	t.Signature = bf.Next(NETWORK_KEY_SIZE)
	t.Payload = bf.Next(int(t.Header.PayloadLength))
	t.From = bf.Next(IP_SIZE)

	return bf.Next(math.MaxInt64), nil
}

type TransactionSlice []Transaction

func (slice TransactionSlice) Len() int {
	return len(slice)
}

func (slice TransactionSlice) Exists(tr Transaction) bool {
	for _, t := range slice {
		if reflect.DeepEqual(t, tr) {
			return true
		}
	}

	return false
}

func (slice TransactionSlice) AddTransaction(tr Transaction) TransactionSlice {
	for i, t := range slice {
		if t.Header.Timestamp > tr.Header.Timestamp {
			slice = append(append(slice[:i], tr), slice[i:]...)
		}
	}
	return append(slice, tr)
}

func (slice *TransactionSlice) MarshalBinary() ([]byte, error) {
	bf := &bytes.Buffer{}

	for _, t := range *slice {
		bs, err := t.MarshalBinary()
		if err != nil {
			return nil, err
		}
		bf.Write(bs)
	}

	return bf.Bytes(), nil
}

func (slice *TransactionSlice) UnMarshalBinary(d []byte) error {
	remaining := d

	for len(remaining) > TRANSACTION_HEADER_SIZE+NETWORK_KEY_SIZE {
		t := new(Transaction)
		rem, err := t.UnMarshalBinary(remaining)

		if err != nil {
			return err
		}
		(*slice) = append((*slice), *t)
		remaining = rem
	}
	return nil
}
