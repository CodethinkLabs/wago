// store is a raft-backed decimalAmount store
// that associates ED25519 keys with a wallet
// currencies that users can exchange with
// each other

package wallet

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/CodethinkLabs/wago/pkg/util"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"golang.org/x/crypto/ed25519"
)

// DecimalAmount represents some decimal currency without
// using floating point
type DecimalAmount struct {
	Value   int64
	Decimal int8
}

func (d DecimalAmount) String() string {
	return fmt.Sprintf("%d.%.2d", d.Value, d.Decimal)
}

// Currency represents some currency string
type Currency string

// Account maps a range of currencies to a decimal amount
type Account map[Currency]DecimalAmount

// Store maps public keys to accounts, representing
// the current state of the state machine
type Store struct {
	proposeC    chan<- string // channel for proposing updates
	mu          sync.RWMutex
	WalletStore map[[ed25519.PublicKeySize]byte]Account // current wallets
	snapshotter *snap.Snapshotter
}

// Transaction is a state change transferring an
// amount of some currency between wallets.
type Transaction struct {
	Src    ed25519.PublicKey
	Dest   ed25519.PublicKey
	Sig    [ed25519.SignatureSize]byte
	Curr   Currency
	Amount DecimalAmount
	Create bool
}

// NewStore creates a new Store to hold currencies
func NewStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error, wg *sync.WaitGroup) *Store {
	s := &Store{proposeC: proposeC, WalletStore: make(map[[ed25519.PublicKeySize]byte]Account), snapshotter: snapshotter}
	s.readCommits(commitC, errorC)
	go func() {
		s.readCommits(commitC, errorC)
		wg.Done()
	}()
	return s
}

// Lookup retrieves the currencies for a given public key
func (s *Store) Lookup(key ed25519.PublicKey) (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.WalletStore[util.ToBytes(key)]
	return v, ok
}

// Search gets a unique public key from the store,
// returning (key, true) if there is exactly
// one match, else (nil, false).
func (s *Store) Search(key string) (ed25519.PublicKey, bool) {
	matches := make([]ed25519.PublicKey, 0)

	for storeKey := range s.WalletStore {
		storeKeyString := hex.EncodeToString(storeKey[:])
		if strings.HasPrefix(storeKeyString, key) {
			match := make(ed25519.PublicKey, len(storeKey))
			copy(match, ed25519.PublicKey(storeKey[:]))
			matches = append(matches, match)
		}
	}

	if len(matches) == 1 {
		return matches[0], true
	}

	return nil, false
}

// Propose when provided a destination wallet, a decimalAmount,
// a currency, and a valid signature on the transaction,
// requests a transfer from the src to the dest of the
// requested decimalAmount.
//
// Performs a simple crypto check to make sure the transaction is
// signed by the src address.
func (s *Store) Propose(trans Transaction) error {
	if !trans.IsVerified() {
		return fmt.Errorf("provided signature does not match the public key")
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(trans); err != nil {
		return err
	}

	s.proposeC <- buf.String()
	return nil
}

// for each previous commit in the channel,
// apply it to the local state machine
func (s *Store) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := s.snapshotter.Load()
			if err == snap.ErrNoSnapshot {
				return
			}
			if err != nil {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}

		var nextTrans Transaction
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&nextTrans); err != nil {
			log.Fatalf("could not decode message (%v)", err)
		}
		if nextTrans.Curr == "" {
			log.Printf("Dropping transaction with invalid currency")
			continue
		}

		if !nextTrans.Create {
			if !nextTrans.IsVerified() {
				log.Printf("Dropping transaction with bad signature")
				// discard transactions with invalid signatures
				continue
			}
			if err := s.CheckBalance(nextTrans.Src, nextTrans.Curr, nextTrans.Amount); err != nil {
				// discard transactions without enough funds
				log.Printf("Dropping transaction with bad balance: %s\n", err)
				continue
			}
		}

		s.mu.Lock()
		transfer := Account{nextTrans.Curr: nextTrans.Amount}
		destWallet := s.WalletStore[util.ToBytes(nextTrans.Dest)]
		if !nextTrans.Create {
			srcWallet := s.WalletStore[util.ToBytes(nextTrans.Src)]
			s.WalletStore[util.ToBytes(nextTrans.Src)] = srcWallet.Subtract(transfer)
		}
		s.WalletStore[util.ToBytes(nextTrans.Dest)] = destWallet.Subtract(transfer.Inverse())
		s.mu.Unlock()

		if walletFile, err := ReadWallet(); err == nil {
			if walletFile.Lookup(nextTrans.Src) != nil {
				fmt.Printf("Transfer of %v %s successfully sent to %x\n", nextTrans.Amount, nextTrans.Curr, nextTrans.Dest)
			}
			if walletFile.Lookup(nextTrans.Dest) != nil {
				if nextTrans.Create {
					fmt.Printf("You got money! Created %v %s successfully\n", nextTrans.Amount, nextTrans.Curr)
				} else {
					fmt.Printf("You got money! Transfer of %v %s successfully received from %x\n", nextTrans.Amount, nextTrans.Curr, nextTrans.Src)
				}
			}
		}
	}

	// after loading all the commits, we check the error stream
	// if all goes well, we should expect errorSent to be false
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

// GetSnapshot marshals the store to be used by the snapshotter
func (s *Store) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.WalletStore)
}

func (s *Store) recoverFromSnapshot(snapshot []byte) error {
	var store map[[ed25519.PublicKeySize]byte]Account
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.WalletStore = store
	return nil
}
