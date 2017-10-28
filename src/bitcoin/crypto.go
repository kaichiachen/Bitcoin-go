package bitcoin

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"log"
	"math/big"
)

type Keypair struct {
	Public  []byte
	Private []byte
}

func GenerateNewKeypair() *Keypair {
	pk, _ := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)

	public := append(pk.PublicKey.X.Bytes(), pk.PublicKey.Y.Bytes()...)
	private := pk.D.Bytes()
	kp := Keypair{Public: public, Private: private}

	log.Println("Public key: ", kp.Public)
	log.Println("Private key: ", kp.Private)
	return &kp
}

func (k *Keypair) Sign(hash []byte) ([]byte, error) {
	d := new(big.Int)
	d.SetBytes(k.Private)

	pk := new(big.Int)
	pk.SetBytes(k.Public)

	pub := splitKey(pk, 2)
	x, y := pub[0], pub[1]

	key := ecdsa.PrivateKey{ecdsa.PublicKey{elliptic.P224(), x, y}, d}

	r, s, _ := ecdsa.Sign(rand.Reader, &key, hash)

	return append(r.Bytes(), s.Bytes()...), nil
}

func SignatureVerify(publicKey, sig, hash []byte) bool {
	b := new(big.Int)
	b.SetBytes(sig)
	bSplit := splitKey(b, 2)
	r, s := bSplit[0], bSplit[1]

	b = new(big.Int)
	b.SetBytes(publicKey)
	pkSplit := splitKey(b, 2)
	x, y := pkSplit[0], pkSplit[1]

	pub := ecdsa.PublicKey{elliptic.P224(), x, y}

	return ecdsa.Verify(&pub, hash, r, s)
}

func splitKey(n *big.Int, parts int) []*big.Int {
	bs := n.Bytes()
	if len(bs)%2 == 1 {
		bs = append([]byte{0}, bs...)
	}

	l := len(bs) / parts
	as := make([]*big.Int, parts)

	for i, _ := range as {
		as[i] = new(big.Int).SetBytes(bs[l*i : l*(i+1)])
	}
	return as
}
