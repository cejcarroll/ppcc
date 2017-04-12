package lib

import (
    "bytes"
    "errors"
    "crypto/cipher"
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


// A basic, verifiable signature
type basicSig struct {
    C abstract.Scalar // challenge
    R abstract.Scalar // response
}

// Returns a secret that depends on on a message and a point
func hashSchnorr(suite abstract.Suite, message []byte, p abstract.Point) abstract.Scalar {
    pb, _ := p.MarshalBinary()
    c := suite.Cipher(pb)
    c.Message(nil, nil, message)
    return suite.Scalar().Pick(c)
}

// This simplified implementation of Schnorr Signatures is based on
// crypto/anon/sig.go
// The ring structure is removed and
// The anonimity set is reduced to one public key = no anonimity
func SchnorrSign(suite abstract.Suite, random cipher.Stream, message []byte,
    privateKey abstract.Scalar) []byte {

    // Create random secret v and public point commitment T
    v := suite.Scalar().Pick(random)
    T := suite.Point().Mul(nil, v)

    // Create challenge c based on message and T
    c := hashSchnorr(suite, message, T)

    // Compute response r = v - x*c
    r := suite.Scalar()
    r.Mul(privateKey, c).Sub(v, r)

    // Return verifiable signature {c, r}
    // Verifier will be able to compute v = r + x*c
    // And check that hashElgamal for T and the message == c
    buf := bytes.Buffer{}
    sig := basicSig{c, r}
    suite.Write(&buf, &sig)
    return buf.Bytes()
}

func SchnorrVerify(suite abstract.Suite, message []byte, publicKey abstract.Point,
    signatureBuffer []byte) error {

    // Decode the signature
    buf := bytes.NewBuffer(signatureBuffer)
    sig := basicSig{}
    if err := suite.Read(buf, &sig); err != nil {
        return err
    }
    r := sig.R
    c := sig.C

    // Compute base**(r + x*c) == T
    var P, T abstract.Point
    P = suite.Point()
    T = suite.Point()
    T.Add(T.Mul(nil, r), P.Mul(publicKey, c))

    // Verify that the hash based on the message and T
    // matches the challange c from the signature
    c = hashSchnorr(suite, message, T)
    if !c.Equal(sig.C) {
        return errors.New("invalid signature")
    }

    return nil
}
