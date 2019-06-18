package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"strings"
	"syscall"
)

const WALLET = "wallet"

type walletentry struct {
	PrivateKey ed25519.PrivateKey `yaml:",flow"`
	PublicKey  ed25519.PublicKey  `yaml:",flow"`
}
type walletfile []walletentry

var password = "123"

func (w walletfile) Lookup(key ed25519.PublicKey) ed25519.PrivateKey {
	for _, entry := range w {
		if bytes.Equal(entry.PublicKey, key) {
			return entry.PrivateKey
		}
	}

	return nil
}

// gets a unique public key from the store by its prefix
func (w walletfile) PrefixSearch(key string) (ed25519.PublicKey, bool) {
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
	} else {
		return nil, false
	}
}

// writes the walletfile to disk
func WriteWallet(wallet walletfile) {
	data, err := yaml.Marshal(&wallet)
	if err != nil {
		log.Fatal(err)
	}

	if password != "" {
		data = AESEncrypt(data, password)
	}

	err = ioutil.WriteFile(WALLET, data, 0600)
	if err != nil {
		log.Fatal("Could not write wallet")
	}
}

// reads the walletfile from disk
func ReadWallet() walletfile {
	data, err := ioutil.ReadFile(WALLET)
	if err != nil {
		data = createWalletFile(data)
	}
	if strings.HasPrefix(string(data), "AESENCRYPT") {
		if password == "" {
			panic("No password, please authenticate first.")
		}

		data = AESDecrypt(data, password)
	}

	var wallet walletfile
	err = yaml.Unmarshal(data, &wallet)
	if err != nil {
		log.Fatal(err)
	}
	return wallet
}

func Authenticate(newPassword *string) {
	if newPassword == nil {
		fmt.Print("Enter Password: ")
		bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
		if err == nil {
			fmt.Println("\nPassword typed: " + string(bytePassword))
		}
		password = strings.TrimSpace(string(bytePassword))
	} else {
		password = *newPassword
	}
}

func GenerateKeys(args []string) (ed25519.PrivateKey, ed25519.PublicKey) {
	var privateKey ed25519.PrivateKey
	var publicKey ed25519.PublicKey
	var err error

	switch len(args) {
	case 0:
		publicKey, privateKey, err = ed25519.GenerateKey(nil)
	case 1:
		publicKey, privateKey, err = ed25519.GenerateKey(bytes.NewReader(hashPassword(args[0])))
	default:
		log.Fatal("Command only supports optional password arg")
	}

	if err != nil {
		log.Fatal("Bad key")
	}

	fmt.Printf("Generating new key: %x\n", publicKey[:6])
	return privateKey, publicKey
}

func AddKeys(wallet walletfile, publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey) walletfile {
	for _, entry := range wallet {
		if string(entry.PublicKey) == string(publicKey) || string(entry.PrivateKey) == string(privateKey) {
			return wallet
		}
	}
	wallet = append(wallet, walletentry{PublicKey: publicKey, PrivateKey: privateKey})
	log.Println("New wallet appended to file")
	return wallet
}

// given some plaintext and a password,
// returns the AES ciphertext^nonce
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

// given some ciphertext, a password, and a nonce
// returns the AES plaintext
func AESDecrypt(data []byte, password string) []byte {
	// strip magic sequence
	data = data[10:]
	if len(data) == 0 {
		return data
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
	log.Printf("ciphertext: %x\n", ciphertext)
	log.Printf("nonce: %x\n", nonce)
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	log.Println("Wallet decrypted")
	return plaintext
}

func createWalletFile(data []byte) []byte {
	// file does not exist
	data = []byte("AESENCRYPT")
	err := ioutil.WriteFile(WALLET, data, 0600)

	if err != nil {
		log.Fatal("Could not create wallet file")
	} else {
		fmt.Println("No wallet found, creating new one.")
	}
	if password == "" {
		fmt.Println("Would you like to set a password? (blank for none)")
		Authenticate(nil)
	}

	return data
}

func hashPassword(password string) []byte {
	return pbkdf2.Key([]byte(password), []byte("deadbeef12345678"), 4096, 32, sha1.New)
}
