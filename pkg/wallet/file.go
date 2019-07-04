// Contains functions for creating, reading, updating, and deleting
// public and private ed25519 key-pairs (credentials) from an encrypted file

package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"syscall"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/sha3"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"
)

// FILENAME is where the wallet file is stored
const FILENAME = "wallet"

var password string

type credentials struct {
	PrivateKey ed25519.PrivateKey `yaml:",flow"`
	PublicKey  ed25519.PublicKey  `yaml:",flow"`
}
type localWallet []credentials

// looks up a public key in the wallet, retrieving the
// private key if one exists, or nil
func (w localWallet) Lookup(key ed25519.PublicKey) ed25519.PrivateKey {
	for _, entry := range w {
		if bytes.Equal(entry.PublicKey, key) {
			return entry.PrivateKey
		}
	}

	return nil
}

// searches for a key in the localWallet starting with the given string
// in the case of ambiguity the function returns (nil, false)
func (w localWallet) PrefixSearch(key string) (ed25519.PublicKey, bool) {
	matches := make([]ed25519.PublicKey, 0)

	for _, storeKey := range w {
		storeKeyString := hex.EncodeToString(storeKey.PublicKey[:])
		if strings.HasPrefix(storeKeyString, key) {
			match := make(ed25519.PublicKey, len(storeKey.PublicKey))
			copy(match, storeKey.PublicKey)
			matches = append(matches, match)
		}
	}

	if len(matches) == 1 {
		return matches[0], true
	}

	return nil, false
}

// gets all the public keys from the wallet
func (w localWallet) GetKeys() []ed25519.PublicKey {
	keys := make([]ed25519.PublicKey, 0)
	for _, key := range w {
		keys = append(keys, key.PublicKey)
	}
	return keys
}

// WriteWallet writes the localWallet to disk
func WriteWallet(wallet localWallet) {
	data, err := yaml.Marshal(&wallet)
	if err != nil {
		log.Fatal(err)
	}

	if password != "" {
		data = AESEncrypt(data, password)
	}

	err = ioutil.WriteFile(FILENAME, data, 0600)
	if err != nil {
		log.Fatal("Could not write wallet")
	}
}

// ReadWallet reads the localWallet from disk
func ReadWallet() (localWallet, error) {
	data, err := ioutil.ReadFile(FILENAME)
	if err != nil {
		data = createWalletFile()
	}
	if strings.HasPrefix(string(data), "AESENCRYPT") {
		data, err = AESDecrypt(data, password)
		if err != nil {
			return nil, fmt.Errorf("invalid password, please run the auth command to set it")
		}
	}

	var wallet localWallet
	if err = yaml.Unmarshal(data, &wallet); err != nil {
		return nil, fmt.Errorf("corrupt or malformed file")
	}
	return wallet, nil
}

// Authenticate stores the provided password for use
// when reading and writing to the encrypted wallet
func Authenticate(newPassword *string) {
	if newPassword == nil {
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			password = strings.TrimSpace(string(bytePassword))
			fmt.Print("\n")
		} else {
			fmt.Println("Error while setting password")
		}
	} else {
		password = *newPassword
	}
}

// GenerateKeys generates an ed25519 keypair from
// an optional seed
func GenerateKeys(seed *string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	var err error

	if seed == nil {
		publicKey, privateKey, err = ed25519.GenerateKey(nil)
	} else {
		publicKey, privateKey, err = ed25519.GenerateKey(bytes.NewReader(hashPassword(*seed)))
	}
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// AddKeyPair appends a given public and private key to
// the localWallet, ignoring them if they already exist
func AddKeyPair(wallet localWallet, publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) localWallet {
	for _, entry := range wallet {
		if string(entry.PublicKey) == string(publicKey) || string(entry.PrivateKey) == string(privateKey) {
			return wallet
		}
	}
	wallet = append(wallet, credentials{PublicKey: publicKey, PrivateKey: privateKey})
	log.Println("New credentials appended to file")
	return wallet
}

// AESEncrypt when given some plaintext and a password,
// returns the AES ciphertext concatted with the nonce
func AESEncrypt(plaintext []byte, password string) []byte {
	passwordHash := hashPassword(password)

	block, err := aes.NewCipher(passwordHash)
	if err != nil {
		panic(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	log.Printf("ciphertext: %x\n", ciphertext)
	log.Printf("nonce: %x\n", nonce)
	ciphertext = append(ciphertext, nonce...) // suffix with nonce
	log.Printf("ciphertext: %x\n", ciphertext[:len(ciphertext)-len(nonce)])
	log.Printf("nonce: %x\n", ciphertext[len(ciphertext)-len(nonce):])

	ciphertext = append([]byte("AESENCRYPT"), ciphertext...) // prefix with magic id

	log.Printf("data out: %x\n", ciphertext)

	log.Println("Wallet encrypted")
	return ciphertext
}

// AESDecrypt when given some ciphertext concatted nonce
// and a password, returns the AES plaintext
func AESDecrypt(data []byte, password string) ([]byte, error) {
	// strip magic sequence
	data = data[10:]
	if len(data) == 0 {
		return data, nil
	}

	passwordHash := hashPassword(password)
	block, err := aes.NewCipher(passwordHash)
	if err != nil {
		panic(err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	cipherSize := len(data) - gcm.NonceSize()
	ciphertext, nonce := data[:cipherSize], data[cipherSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	log.Println("Wallet decrypted")
	return plaintext, nil
}

// creates a new, empty wallet file
func createWalletFile() []byte {
	fmt.Println("Creating new wallet file.")
	if password == "" {
		fmt.Println("Would you like to set a password? (blank for none)")
		Authenticate(nil)
	}

	var data []byte
	if password != "" {
		data = []byte("AESENCRYPT") // encrypted
	} else {
		data = []byte("") // not encrypted
	}

	err := ioutil.WriteFile(FILENAME, data, 0600)
	if err != nil {
		log.Fatal("Could not create wallet file")
	}

	return data
}

// hashes the provided user password into an appropriate AES key
// takes the sha3 of the user's password as the salt for pbkdf2.
func hashPassword(password string) []byte {
	saltHasher := sha3.New256()
	saltHasher.Write([]byte(password))

	return pbkdf2.Key([]byte(password), saltHasher.Sum(nil), 4096, 32, sha3.New256)
}
