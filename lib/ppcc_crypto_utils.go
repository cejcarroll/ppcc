package lib

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/crypto.v0/random"
	//"math/rand"
	//"sync"
)

type PPCC struct {
    suite           abstract.Suite
    publics         []abstract.Point
    private         abstract.Scalar
    signKey         abstract.Scalar
    VerifyKey       abstract.Point
}

func NewPPCC(suite abstract.Suite, private abstract.Scalar, publics []abstract.Point) *PPCC {
    ppcc := &PPCC{
        suite:      suite,
        private:    private,
        publics:    publics,
    }
    ppcc.createSigKeys()
    return ppcc;
}

func (c *PPCC) EncryptTelecomMessage(message string, idx int) (
    K abstract.Point, C abstract.Point, remainder []byte) {

    return ElGamalEncrypt(c.suite, c.publics[idx], []byte(message))
}

func (c *PPCC) DecryptTelecomMessage(K, C abstract.Point) (message string, err error){
    bytes, e := ElGamalDecrypt(c.suite, c.private, K, C)
    message = string(bytes)
    err = e
    return
}

func (c *PPCC) createSigKeys() {
    c.signKey = c.suite.Scalar().Pick(random.Stream)
    c.VerifyKey = c.suite.Point().Mul(nil, c.signKey)
}

func (c *PPCC) SignMessage (message string) []byte {
    return SchnorrSign(c.suite, random.Stream, []byte(message), c.signKey)
}

func (c *PPCC) VerifyMessage (message string, pubKey abstract.Point, sigBuffer []byte) error {
    return SchnorrVerify(c.suite, []byte(message), pubKey, sigBuffer)
}
