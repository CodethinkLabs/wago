// walletstore is a raft-backed decimalAmount store
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
	"github.com/CodethinkLabs/wago/pkg/util"
	"go.etcd.io/etcd/etcdserver/api/snap"
	"golang.org/x/crypto/ed25519"
	"log"
	"strings"
	"sync"
)

type DecimalAmount struct {
	Value   int64
	Decimal int8
}

type Currency string
type Currencies map[Currency]DecimalAmount

// a key-value store backed by raft
type WalletStore struct {
	proposeC    chan<- string // channel for proposing updates
	mu          sync.RWMutex
	WalletStore map[[ed25519.PublicKeySize]byte]Currencies // current wallets
	snapshotter *snap.Snapshotter
}

type transaction struct {
	Src    ed25519.PublicKey
	Dest   ed25519.PublicKey
	Sig    [ed25519.SignatureSize]byte
	Curr   Currency
	Amount DecimalAmount
	Create bool
}

func NewWalletStore(snapshotter *snap.Snapshotter, proposeC chan<- string, commitC <-chan *string, errorC <-chan error) *WalletStore {
	s := &WalletStore{proposeC: proposeC, WalletStore: make(map[[32]byte]Currencies), snapshotter: snapshotter}
	s.readCommits(commitC, errorC)
	go s.readCommits(commitC, errorC)
	return s
}

// retrieves the currencies for a given public key
func (s *WalletStore) Lookup(key ed25519.PublicKey) (Currencies, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.WalletStore[util.ToBytes(key)]
	return v, ok
}

// gets a unique public key from the store by its prefix
func (s *WalletStore) PrefixSearch(key string) (ed25519.PublicKey, bool) {
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
	} else {
		return nil, false
	}
}

// Provided a destination wallet, a decimalAmount (identifier and volume),
// and a valid signature on the amount, this requests a transfer
// from the src to the dest of the requested decimalAmount
//
// performs a simple crypto check to make sure the transaction is
// signed by the src address
func (s *WalletStore) Propose(trans transaction) error {
	log.Printf("prop signature: %x\n", trans.Sig)
	if !trans.IsVerified() {
		return fmt.Errorf("invalid signature")
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(trans); err != nil {
		log.Fatal(err)
	}
	s.proposeC <- buf.String()

	return nil
}

// for each previous commit in the channel,
// apply it to the local state machine
func (s *WalletStore) readCommits(commitC <-chan *string, errorC <-chan error) {
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

		var nextTrans transaction
		dec := gob.NewDecoder(bytes.NewBufferString(*data))
		if err := dec.Decode(&nextTrans); err != nil {
			log.Fatalf("raftrade: could not decode message (%v)", err)
		}

		if !nextTrans.Create {
			if !nextTrans.IsVerified() {
				fmt.Println("Dropping transaction with bad signature")
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
		transfer := Currencies{nextTrans.Curr: nextTrans.Amount}
		destWallet := s.WalletStore[util.ToBytes(nextTrans.Dest)]
		if !nextTrans.Create {
			srcWallet := s.WalletStore[util.ToBytes(nextTrans.Src)]
			s.WalletStore[util.ToBytes(nextTrans.Src)] = srcWallet.Subtract(transfer)
		}
		s.WalletStore[util.ToBytes(nextTrans.Dest)] = destWallet.Subtract(transfer.Inverse())
		s.mu.Unlock()

		walletFile := ReadWallet()
		if walletFile.Lookup(nextTrans.Src) != nil {
			fmt.Printf("\nTransfer of %v %s successfully sent to %x\n", nextTrans.Amount, nextTrans.Curr, nextTrans.Dest)
		}
		if walletFile.Lookup(nextTrans.Dest) != nil {
			fmt.Printf("\nYou got money! Transfer of %v %s successfully received from %x\n", nextTrans.Amount, nextTrans.Curr, nextTrans.Src)
		}
	}

	// after loading all the commits, we check the error stream
	// if all goes well, we should expect errorSent to be false
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *WalletStore) GetSnapshot() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.WalletStore)
}

func (s *WalletStore) recoverFromSnapshot(snapshot []byte) error {
	var store map[[ed25519.PublicKeySize]byte]Currencies
	if err := json.Unmarshal(snapshot, &store); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.WalletStore = store
	return nil
}
