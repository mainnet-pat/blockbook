//go:build unittest

package api

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/bchain/coins/bch"
	"github.com/trezor/blockbook/bchain/coins/btc"
	"github.com/trezor/blockbook/common"
	"github.com/trezor/blockbook/db"
	"github.com/trezor/blockbook/fiat"
)

func TestSystemInfoInSync(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	oldStart := now.Add(-time.Minute)

	tests := []struct {
		name          string
		inSync        bool
		initialSync   bool
		chainType     bchain.ChainType
		bestHeight    uint32
		backendBlocks int
		lastBlockTime time.Time
		startSync     time.Time
		blockPeriod   time.Duration
		want          bool
	}{
		{
			name:          "reports evm synced when active sync loop is already at backend tip",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
			want:          true,
		},
		{
			name:          "does not hide stale evm tip when heights match",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-25 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
		{
			name:          "does not report synced while local height is behind",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    90,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
		{
			name:          "reports evm synced within one block of a fresh tip",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    99,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
			want:          true,
		},
		{
			name:          "reports evm synced on a sub-second chain at the tip",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-1 * time.Second),
			startSync:     oldStart,
			blockPeriod:   250 * time.Millisecond,
			want:          true,
		},
		{
			name:          "does not report synced more than one block behind tip",
			chainType:     bchain.ChainEthereumType,
			bestHeight:    98,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
		{
			name:          "does not report synced during initial sync",
			initialSync:   true,
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
		{
			name:          "keeps startup grace for fresh regular sync",
			chainType:     bchain.ChainBitcoinType,
			backendBlocks: 100,
			startSync:     now.Add(-2 * time.Second),
			want:          true,
		},
		{
			name:          "marks already synced evm stale",
			inSync:        true,
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-25 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
		{
			name:          "keeps already synced evm fresh",
			inSync:        true,
			chainType:     bchain.ChainEthereumType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
			want:          true,
		},
		{
			name:          "does not extend tip equality rescue to bitcoin",
			chainType:     bchain.ChainBitcoinType,
			bestHeight:    100,
			backendBlocks: 100,
			lastBlockTime: now.Add(-10 * time.Second),
			startSync:     oldStart,
			blockPeriod:   2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := systemInfoInSync(tt.inSync, tt.initialSync, tt.chainType, tt.bestHeight, tt.backendBlocks, tt.lastBlockTime, tt.startSync, now, tt.blockPeriod)
			if got != tt.want {
				t.Fatalf("systemInfoInSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSecondaryTicker_SkipsLookupWithoutSecondaryCurrency(t *testing.T) {
	w := &Worker{
		fiatRates: &fiat.FiatRates{Enabled: true},
	}
	originalGetter := getCurrentTicker
	defer func() {
		getCurrentTicker = originalGetter
	}()

	calls := 0
	getCurrentTicker = func(_ *fiat.FiatRates, _, _ string) *common.CurrencyRatesTicker {
		calls++
		return &common.CurrencyRatesTicker{}
	}

	ticker := w.getSecondaryTicker("")
	if ticker != nil {
		t.Fatalf("expected nil ticker when secondary currency is not requested, got %+v", ticker)
	}
	if calls != 0 {
		t.Fatalf("expected no ticker lookup call, got %d", calls)
	}
}

func TestGetSecondaryTicker_PerformsLookupWithSecondaryCurrency(t *testing.T) {
	w := &Worker{
		fiatRates: &fiat.FiatRates{Enabled: true},
	}
	originalGetter := getCurrentTicker
	defer func() {
		getCurrentTicker = originalGetter
	}()

	calls := 0
	expected := &common.CurrencyRatesTicker{Rates: map[string]float32{"usd": 1}}
	getCurrentTicker = func(_ *fiat.FiatRates, _, _ string) *common.CurrencyRatesTicker {
		calls++
		return expected
	}

	ticker := w.getSecondaryTicker("usd")
	if ticker != expected {
		t.Fatalf("unexpected ticker returned: got %+v, want %+v", ticker, expected)
	}
	if calls != 1 {
		t.Fatalf("expected one ticker lookup call, got %d", calls)
	}
}

func TestTronBalanceHistoryOverrides(t *testing.T) {
	tests := []struct {
		name              string
		payload           string
		fallbackAmount    string
		hasFallbackAmount bool
		wantOverride      bool
		wantDirection     tronBalanceHistoryDirection
		wantAmount        string
	}{
		{
			name:              "freeze uses stake amount",
			payload:           `{"operation":"freeze","stakeAmount":"42000000"}`,
			fallbackAmount:    "1",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionOutgoing,
			wantAmount:        "42000000",
		},
		{
			name:              "withdraw uses unstake amount",
			payload:           `{"operation":"withdraw","unstakeAmount":"77000000"}`,
			fallbackAmount:    "1",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionIncoming,
			wantAmount:        "77000000",
		},
		{
			name:              "withdraw falls back to tx value",
			payload:           `{"operation":"withdraw"}`,
			fallbackAmount:    "123",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionIncoming,
			wantAmount:        "123",
		},
		{
			name:              "vote reward amount uses claimed vote reward",
			payload:           `{"operation":"voteRewardAmount","claimedVoteReward":"6500000"}`,
			fallbackAmount:    "1",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionIncoming,
			wantAmount:        "6500000",
		},
		{
			name:              "vote reward amount falls back to tx value",
			payload:           `{"operation":"voteRewardAmount"}`,
			fallbackAmount:    "321",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionIncoming,
			wantAmount:        "321",
		},
		{
			name:              "freeze invalid amount falls back to tx value",
			payload:           `{"operation":"freeze","stakeAmount":"not-a-number"}`,
			fallbackAmount:    "999",
			hasFallbackAmount: true,
			wantOverride:      true,
			wantDirection:     tronBalanceHistoryDirectionOutgoing,
			wantAmount:        "999",
		},
		{
			name:          "unfreeze has explicit no-move override",
			payload:       `{"operation":"unfreeze","unstakeAmount":"77000000"}`,
			wantOverride:  true,
			wantDirection: tronBalanceHistoryDirectionNone,
			wantAmount:    "0",
		},
		{
			name:         "non-freeze operation has no override",
			payload:      `{"operation":"transfer","stakeAmount":"42000000"}`,
			wantOverride: false,
		},
		{
			name:         "invalid json has no override",
			payload:      `{`,
			wantOverride: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fallback *big.Int
			if tt.hasFallbackAmount {
				var ok bool
				fallback, ok = new(big.Int).SetString(tt.fallbackAmount, 10)
				if !ok {
					t.Fatalf("invalid fallback amount in test: %q", tt.fallbackAmount)
				}
			}

			override, hasOverride := tronBalanceHistoryOverrideFromExtraData(json.RawMessage(tt.payload), fallback)
			if hasOverride != tt.wantOverride {
				t.Fatalf("override mismatch: got %v want %v", hasOverride, tt.wantOverride)
			}
			if !tt.wantOverride {
				return
			}
			if override.direction != tt.wantDirection {
				t.Fatalf("direction mismatch: got %v want %v", override.direction, tt.wantDirection)
			}
			if got := override.amount.String(); got != tt.wantAmount {
				t.Fatalf("amount mismatch: got %s want %s", got, tt.wantAmount)
			}
		})
	}
}

func TestTxFromTxAddress_PreservesBcashTokenAndScriptAddress(t *testing.T) {
	parser, err := bch.NewBCashParser(bch.GetChainParams("test"), &btc.Configuration{AddressFormat: "cashaddr"})
	if err != nil {
		t.Fatalf("NewBCashParser() error = %v", err)
	}
	rocks, err := db.NewRocksDB(t.TempDir(), 1, 1, parser, &common.Metrics{}, false)
	if err != nil {
		t.Fatalf("NewRocksDB() error = %v", err)
	}
	defer rocks.Close()

	addrDesc := bchain.AddressDescriptor{0xaa, 0xbb, 0xcc}
	token := &bchain.BcashToken{
		Category: bytes.Repeat([]byte{0x11}, 32),
		Amount:   common.Amount(*big.NewInt(42)),
		Nft: &bchain.BcashTokenNft{
			Capability: bchain.NFTCapabilityLabelMutable,
			Commitment: []byte{0xde, 0xad, 0xbe, 0xef},
		},
	}
	ta := &db.TxAddresses{
		Height: 313026,
		Outputs: []db.TxOutput{
			{
				AddrDesc:   addrDesc,
				ValueSat:   *big.NewInt(12345),
				BcashToken: token,
			},
		},
	}
	w := &Worker{
		chainParser: parser,
		db:          rocks,
		is:          &common.InternalState{CoinShortcut: "BCH"},
	}

	tx := w.txFromTxAddress("txid", ta, &db.BlockInfo{Hash: "blockhash", Time: 1700000000}, 313026, nil)
	if tx == nil {
		t.Fatal("txFromTxAddress() returned nil")
	}
	if len(tx.Vout) != 1 {
		t.Fatalf("unexpected vout count: got %d, want 1", len(tx.Vout))
	}
	if got := tx.Vout[0].Addresses; len(got) != 1 || got[0] != "script-aabbcc" {
		t.Fatalf("unexpected output addresses: got %+v, want [script-aabbcc]", got)
	}
	if tx.Vout[0].BcashToken == nil {
		t.Fatal("expected output token data")
	}
	if got, want := tx.Vout[0].BcashToken.Category, "1111111111111111111111111111111111111111111111111111111111111111"; got != want {
		t.Fatalf("unexpected token category: got %s, want %s", got, want)
	}
	if got, want := tx.Vout[0].BcashToken.Amount.String(), "42"; got != want {
		t.Fatalf("unexpected token amount: got %s, want %s", got, want)
	}
	if tx.Vout[0].BcashToken.Nft == nil {
		t.Fatal("expected nft token data")
	}
	if got, want := tx.Vout[0].BcashToken.Nft.Capability, "mutable"; got != want {
		t.Fatalf("unexpected nft capability: got %s, want %s", got, want)
	}
	if got, want := tx.Vout[0].BcashToken.Nft.Commitment, "deadbeef"; got != want {
		t.Fatalf("unexpected nft commitment: got %s, want %s", got, want)
	}

	bcashSpecific, err := w.bcashPostProcessApiTx(tx.Txid, &tx.Vin, &tx.Vout)
	if err != nil {
		t.Fatalf("bcashPostProcessApiTx() error = %v", err)
	}
	if bcashSpecific == nil || len(bcashSpecific.TokenVouts) != 1 || bcashSpecific.TokenVouts[0] == nil {
		t.Fatalf("unexpected bcashSpecific tokenVouts: %+v", bcashSpecific)
	}
	if got, want := bcashSpecific.TokenVouts[0].Amount.String(), "42"; got != want {
		t.Fatalf("unexpected post-processed token amount: got %s, want %s", got, want)
	}
	if got, want := bcashSpecific.TokenVouts[0].Nft.Commitment, []byte{0xde, 0xad, 0xbe, 0xef}; !bytes.Equal(got, want) {
		t.Fatalf("unexpected post-processed commitment: got %x, want %x", got, want)
	}
}
