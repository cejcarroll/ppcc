package lib

import (
	"encoding/hex"
	"fmt"

	"gopkg.in/dedis/crypto.v0/nist"
)

// Example of using Schnorr
func ExampleSchnorr() {
    // Crypto setup
    suite := nist.NewAES128SHA256P256()
    rand := suite.Cipher([]byte("example"))

    // Create a public/private keypair (X,x)
    x := suite.Scalar().Pick(rand) // create a private key x
    X := suite.Point().Mul(nil, x) // corresponding public key X

    // Generate the signature
    M := []byte("Hello World!") // message we want to sign
    sig := SchnorrSign(suite, rand, M, x)
    fmt.Print("Signature:\n" + hex.Dump(sig))

    // Verify the signature against the correct message
    err := SchnorrVerify(suite, M, X, sig)
    if err != nil {
        panic(err.Error())
    }
    fmt.Println("Signature verified against correct message.")
    // Output:
    // Signature:
    // 00000000  c1 7a 91 74 06 48 5d 53  d4 92 27 71 58 07 eb d5  |.z.t.H]S..'qX...|
    // 00000010  75 a5 89 92 78 67 fc b1  eb 36 55 63 d1 32 12 20  |u...xg...6Uc.2. |
    // 00000020  2c 78 84 81 04 0d 2a a8  fa 80 d0 e8 c3 14 65 e3  |,x....*.......e.|
    // 00000030  7f f2 7c 55 c5 d2 c6 70  51 89 40 cd 63 50 bf c6  |..|U...pQ.@.cP..|
    // Signature verified against correct message.

	println("Schnorr signature verified correctly")
}
