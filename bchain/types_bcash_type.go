package bchain

import "github.com/trezor/blockbook/common"

type BcashNFTCapabilityType uint8

const (
	PREFIX_TOKEN       = 0xef
	HAS_COMMITMENT_LEN = 0x40
	HAS_NFT            = 0x20
	HAS_AMOUNT         = 0x10
)

const (
	NFTCapabilityNone    BcashNFTCapabilityType = 0
	NFTCapabilityMutable BcashNFTCapabilityType = 1
	NFTCapabilityMinting BcashNFTCapabilityType = 2
)

type BcashNFTCapabilityLabel string

// NFTCapabilityLabelToNumber maps a BcashNFTCapabilityLabel to its corresponding BcashNFTCapabilityType number.
func NFTCapabilityLabelToNumber(label BcashNFTCapabilityLabel) BcashNFTCapabilityType {
	switch label {
	case NFTCapabilityLabelNone:
		return NFTCapabilityNone
	case NFTCapabilityLabelMutable:
		return NFTCapabilityMutable
	case NFTCapabilityLabelMinting:
		return NFTCapabilityMinting
	default:
		return NFTCapabilityNone
	}
}

const (
	NFTCapabilityLabelNone    BcashNFTCapabilityLabel = "none"
	NFTCapabilityLabelMutable BcashNFTCapabilityLabel = "mutable"
	NFTCapabilityLabelMinting BcashNFTCapabilityLabel = "minting"
)

// Map from type to label
func ToNFTCapabilityLabel(c BcashNFTCapabilityType) BcashNFTCapabilityLabel {
	switch c {
	case NFTCapabilityNone:
		return NFTCapabilityLabelNone
	case NFTCapabilityMutable:
		return NFTCapabilityLabelMutable
	case NFTCapabilityMinting:
		return NFTCapabilityLabelMinting
	default:
		return NFTCapabilityLabelNone
	}
}

// Map from label to type
func ToNFTCapabilityType(l BcashNFTCapabilityLabel) BcashNFTCapabilityType {
	switch l {
	case NFTCapabilityLabelNone:
		return NFTCapabilityNone
	case NFTCapabilityLabelMutable:
		return NFTCapabilityMutable
	case NFTCapabilityLabelMinting:
		return NFTCapabilityMinting
	default:
		return NFTCapabilityNone
	}
}

type BcashTokenNft struct {
	Capability BcashNFTCapabilityLabel `json:"capability" ts_doc:"Capability of the NFT, which can be 'none', 'mutable', or 'minting'"`
	Commitment string                  `json:"commitment" ts_doc:"Commitment of the NFT, hex encoded, maximum 40 bytes"`
}

// BcashToken represents a CashToken in a BitcoinCash transaction
type BcashToken struct {
	Category string         `json:"category" ts_doc:"Identifier of the token, which is a 32-byte hash of its genesis transaction"`
	Amount   common.Amount  `json:"amount" ts_doc:"Fungible token amount in base units"`
	Nft      *BcashTokenNft `json:"nft,omitempty" ts_doc:"Optional pointer to a BcashTokenNft object if the token also holds an NFT"`
}
