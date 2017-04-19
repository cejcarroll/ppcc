package lib

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/nist"
	//"github.com/lihiid/Crypto/abstract"
	//"github.com/lihiid/Crypto/nist"
	"gopkg.in/dedis/crypto.v0/random"
	"testing"
)

func TestCU(t *testing.T) {

	suite := nist.NewAES128SHA256P256()
	var c0 *PPCC
	var c1 *PPCC
	var c2 *PPCC

	a := suite.Scalar().Pick(random.Stream)
	A := suite.Point().Mul(nil, a)
	b := suite.Scalar().Pick(random.Stream)
	B := suite.Point().Mul(nil, b)
	c := suite.Scalar().Pick(random.Stream)
	C := suite.Point().Mul(nil, c)

	//d := suite.Scalar().Pick(random.Stream)

	publics := []abstract.Point{A, B, C}
	private1 := a
	private2 := b
	private3 := c
	//private4 := d

	c0 = NewPPCC(suite, private1, publics)
	c1 = NewPPCC(suite, private2, publics)
	c2 = NewPPCC(suite, private3, publics)

    message0 := "Test msg"
    K0, C0, _ := c0.EncryptTelecomMessage(message0, 1)
    decoded0, err := c1.DecryptTelecomMessage(K0, C0)
    if err != nil || message0 != decoded0 {
        panic("ERROR: Telecom Decryption failed")
    }

    println("PASS: Telecom Decryption")

    sig0 := c1.SignMessage(decoded0)
    pubKey1 := c1.VerifyKey

    if c2.VerifyMessage(decoded0, pubKey1, sig0) != nil {
        panic("ERROR: Signature Verification failed")
    }

    println("PASS: Telecom Signature")
}

