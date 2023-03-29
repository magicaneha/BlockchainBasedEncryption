/*
@Neha Shecter
*/

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"math/big"

	"github.com/Nik-U/pbc"
)

// Simulates the access policy for the document
var accessPolicy = "foo or (bar and baz)"

// Simulates the original document data
var documentData = []byte("Hello, World!")

func main() {
	// Generate a new pairing
	pairing := pbc.GenerateA(160, 512)

	// Generate a new AES key for symmetric encryption of the private key
	aesKey := make([]byte, 32)
	_, err := rand.Read(aesKey)
	if err != nil {
		panic(err)
	}

	// Generate a new public-private key pair for the verifier using the access policy
	publicKey, privateKey, err := generateKeyPair(pairing, accessPolicy)
	if err != nil {
		panic(err)
	}

	// Simulate the verifier encrypting the document using CP-ABE
	ciphertext, err := encryptDocument(pairing, publicKey, documentData)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Encrypted document: %x\n", ciphertext)

	// Simulate the verifier encrypting the private key using AES-256 with the symmetric key
	encryptedPrivateKey, err := encryptPrivateKey(aesKey, privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Encrypted private key: %x\n", encryptedPrivateKey)

	// Simulate the decrypting the private key using AES-256 with the symmetric key
	decryptedPrivateKey, err := decryptPrivateKey(aesKey, encryptedPrivateKey)
	if err != nil {
		panic(err)
	}

	// Simulate the verifier decrypting the document using CP-ABE and the decrypted private key
	decryptedDocument, err := decryptDocument(pairing, decryptedPrivateKey, ciphertext)
	if err != nil {
		panic(err)
	}
	fmt.Println("Decrypted document:", string(decryptedDocument))
}

func generateKeyPair(pairing *pbc.Pairing, accessPolicy string) (*pbc.Element, *rsa.PrivateKey, error) {
	// Generate a new RSA private key for the verifier
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Generate a new public key for CP-ABE based on the access policy
	policy := pbc.NewPolicy()
	err = policy.UnmarshalText([]byte(accessPolicy))
	if err != nil {
		return nil, nil, err
	}
	publicKey, err := pairing.NewG2().Rand()
	if err != nil {
		return nil, nil, err
	}
	masterKey, err := pairing.NewZr().Rand()
	if err != nil {
		return nil, nil, err
	}
	publicKey, err = publicKey.MulZn(masterKey)
	if err != nil {
		return nil, nil, err
	}
	privateKeyG1 := pairing.NewG1().Set1()
	privateKeyG1, err = privateKeyG1.MulZn(masterKey)
	if err != nil {
		return nil, nil, err
	}
	publicKey, err = publicKey.MapG1(privateKeyG1)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

func encryptDocument(pairing *pbc.Pairing, publicKey *pbc.Element, data []byte) ([]byte, error) {
	// Generate a new symmetric key for AES-256 encryption of the document
	aesKey := make([]byte, 32)
	_, err := rand.Read(aesKey)
	if err != nil {
		return nil, err
	}

	// Encrypt the document using AES-256 with the symmetric key
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	encryptedData := aesGcm.Seal(nil, nonce, data, nil)

	// Generate a new random element r and compute the ABE ciphertext
	r := pairing.NewZr().Rand()
	C1 := pairing.NewG2().Set1()
	C1, err = C1.MulZn(r)
	if err != nil {
		return nil, err
	}
	C2 := pairing.NewG2().SetBytes(encryptedData)
	C2, err = C2.MulZn(pairing.NewZr().SetBytes(aesKey))
	if err != nil {
		return nil, err
	}
	policy := pbc.NewPolicy()
	err = policy.UnmarshalText([]byte(accessPolicy))
	if err != nil {
		return nil, err
	}
	C, err := pairing.EncryptG2(publicKey, policy, []byte{byte(r.Int64())})
	if err != nil {
		return nil, err
	}

	// Serialize the encrypted data and ABE ciphertext as a single byte slice
	encryptedDocument := make([]byte, 0)
	encryptedDocument = append(encryptedDocument, C1.Bytes()...)
	encryptedDocument = append(encryptedDocument, C2...)
	encryptedDocument = append(encryptedDocument, C.Bytes()...)

	return encryptedDocument, nil
}
