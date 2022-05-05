package keyshelper

import (
	"crypto/dsa"
	crand "crypto/rand"
	"encoding/gob"
	"log"
	"math/big"
	"os"
)

type BlockSignature struct {
	R         *big.Int
	S         *big.Int
	ProcessNr int
}
type Signature struct {
	R *big.Int
	S *big.Int
}
type keysPaths struct {
	privateKeyPath string
	publicKeyPath  string
}

var storedPaths keysPaths

func StorePaths(private, public string) {
	storedPaths.privateKeyPath = private
	storedPaths.publicKeyPath = public
}
func OpenPrivate() *dsa.PrivateKey {
	fdprivKey, err := os.Open(storedPaths.privateKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fdprivKey.Close()

	return loadPrivateKey(fdprivKey)
}

func OpenPublic() *dsa.PublicKey {
	fdPublicKey, err := os.Open(storedPaths.publicKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fdPublicKey.Close()

	return loadPublicKey(fdPublicKey)
}

func loadPrivateKey(key *os.File) (DSAPrivateKey *dsa.PrivateKey) {
	decoder := gob.NewDecoder(key)
	err := decoder.Decode(&DSAPrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	return
}
func loadPublicKey(key *os.File) (DSAPublicKey *dsa.PublicKey) {
	decoder := gob.NewDecoder(key)
	err := decoder.Decode(&DSAPublicKey)
	if err != nil {
		log.Fatal(err)
	}
	return
}
func SignBlockMessage(i int, msgHash []byte) BlockSignature {

	DSAPrivateKey := OpenPrivate()

	r, s, err := dsa.Sign(crand.Reader, DSAPrivateKey, msgHash)
	if err != nil {
		log.Fatal(err)
	}
	signature := BlockSignature{
		R: r,
		S: s,
	}
	signature.ProcessNr = i
	return signature
}
func SignMessage(msgHash []byte) Signature {

	DSAPrivateKey := OpenPrivate()

	r, s, err := dsa.Sign(crand.Reader, DSAPrivateKey, msgHash)
	if err != nil {
		log.Fatal(err)
	}
	signature := Signature{
		R: r,
		S: s,
	}
	return signature
}
func GenKey(privateKeyPath, publicKeyPath string) {

	params := new(dsa.Parameters)

	if err := dsa.GenerateParameters(params, crand.Reader, dsa.L1024N160); err != nil {
		log.Fatal(err)
	}

	privatekey := new(dsa.PrivateKey)
	privatekey.PublicKey.Parameters = *params
	dsa.GenerateKey(privatekey, crand.Reader)
	pubkey := privatekey.PublicKey

	privatekeyfile, err := os.Create(privateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	publickeyfile, err := os.Create(publicKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := gob.NewEncoder(privatekeyfile).Encode(privatekey); err != nil {
		log.Fatal(err)
	}

	privatekeyfile.Close()

	if err := gob.NewEncoder(publickeyfile).Encode(pubkey); err != nil {
		log.Fatal(err)
	}

	publickeyfile.Close()
}
func GetPublic() *dsa.PublicKey {
	return OpenPublic()
}
