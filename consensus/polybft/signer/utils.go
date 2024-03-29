package bls

import (
	"crypto/rand"
	"math/big"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/umbracle/ethgo/abi"
)

var (
	addressABIType = abi.MustNewType("address")
	uint256ABIType = abi.MustNewType("uint256")
)

// GenerateBlsKey creates a random private and its corresponding public keys
func GenerateBlsKey() (*PrivateKey, error) {
	s, err := randomK(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{s: s}, nil
}

// CreateRandomBlsKeys creates an array of random private and their corresponding public keys
func CreateRandomBlsKeys(total int) ([]*PrivateKey, error) {
	blsKeys := make([]*PrivateKey, total)

	for i := 0; i < total; i++ {
		blsKey, err := GenerateBlsKey()
		if err != nil {
			return nil, err
		}

		blsKeys[i] = blsKey
	}

	return blsKeys, nil
}

// MarshalMessageToBigInt marshalls message into two big ints
// first we must convert message bytes to point and than for each coordinate we create big int
func MarshalMessageToBigInt(message, domain []byte) ([2]*big.Int, error) {
	point, err := hashToPoint(message, domain)
	if err != nil {
		return [2]*big.Int{}, err
	}

	buf := point.Marshal()

	return [2]*big.Int{
		new(big.Int).SetBytes(buf[0:32]),
		new(big.Int).SetBytes(buf[32:64]),
	}, nil
}

// H_MODIFY: Generate KOSK signature without SupernetManager parameter (not L2 chain)
// MakeKOSKSignature creates KOSK signature which prevents rogue attack
func MakeKOSKSignature(
	privateKey *PrivateKey, address types.Address, chainID int64, domain []byte) (*Signature, error) {
	message, err := abi.Encode(
		[]interface{}{address, big.NewInt(chainID)},
		abi.MustNewType("tuple(address, uint256)"))
	if err != nil {
		return nil, err
	}

	// abi.Encode adds 12 zero bytes before actual address bytes
	return privateKey.Sign(message[12:], domain)
}
