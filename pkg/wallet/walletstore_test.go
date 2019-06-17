// walletstore is a raft-backed decimalAmount store
// that associates ED25519 keys with a wallet
// currencies that users can exchange with
// each other

package wallet

import (
	"encoding/hex"
	"reflect"
	"sync"
	"testing"

	"go.etcd.io/etcd/etcdserver/api/snap"
	"golang.org/x/crypto/ed25519"
)

func Test_walletstore_PrefixSearch(t *testing.T) {
	type fields struct {
		proposeC    chan<- string
		mu          sync.RWMutex
		walletStore map[[ed25519.PublicKeySize]byte]Currencies
		snapshotter *snap.Snapshotter
	}
	type args struct {
		key string
	}

	var goodKey [32]byte
	var badKey [32]byte
	keySlice, _ := hex.DecodeString("deadbeef")
	copy(goodKey[:], keySlice)
	keySlice, _ = hex.DecodeString("1234beef")
	copy(badKey[:], keySlice)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   ed25519.PublicKey
		want1  bool
	}{
		{
			name: "Do smth",
			fields: fields{walletStore: map[[32]byte]Currencies{
				goodKey: map[Currency]DecimalAmount{},
				badKey:  map[Currency]DecimalAmount{},
			}},
			args:  args{key: "deadbeef"},
			want:  ed25519.PublicKey(goodKey[:]),
			want1: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &WalletStore{
				proposeC:    tt.fields.proposeC,
				mu:          tt.fields.mu,
				WalletStore: tt.fields.walletStore,
				snapshotter: tt.fields.snapshotter,
			}
			got, got1 := s.PrefixSearch(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("walletstore.PrefixSearch() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("walletstore.PrefixSearch() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
