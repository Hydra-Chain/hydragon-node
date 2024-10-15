package validator

import (
	"math/big"

	"github.com/0xPolygon/polygon-edge/types"
	"github.com/hashicorp/go-hclog"
)

var InitialMinStake = new(big.Int).Mul(big.NewInt(1e18), big.NewInt(15000))

// ValidatorSet interface of the current validator set
type ValidatorSet interface {
	// Includes check if given address is among the current validator set
	Includes(address types.Address) bool

	// Len returns the size of the validator set
	Len() int

	// Accounts returns the list of the ValidatorMetadata
	Accounts() AccountSet

	// HasQuorum checks if submitted signers have reached quorum
	HasQuorum(blockNumber uint64, signers map[types.Address]struct{}) bool

	// GetVotingPowers retrieves map: string(address) -> vp
	GetVotingPowers() map[string]*big.Int
}

var _ ValidatorSet = (*validatorSet)(nil)

type validatorSet struct {
	// validators represents current list of validators
	validators AccountSet

	// votingPowerMap represents voting powers per validator address
	votingPowerMap map[types.Address]*big.Int

	// totalVotingPower is sum of active validator set
	totalVotingPower *big.Int

	// logger instance
	logger hclog.Logger
}

// NewValidatorSet creates a new validator set.
func NewValidatorSet(valz AccountSet, logger hclog.Logger) *validatorSet {
	votingPowerMap := make(map[types.Address]*big.Int, len(valz))
	for _, val := range valz {
		votingPowerMap[val.Address] = val.VotingPower
	}

	return &validatorSet{
		validators:       valz,
		votingPowerMap:   votingPowerMap,
		totalVotingPower: valz.GetTotalVotingPower(),
		logger:           logger.Named("validator_set"),
	}
}

// HasQuorum determines if there is quorum of enough signers reached,
// based on its voting power and quorum size
func (vs validatorSet) HasQuorum(blockNumber uint64, signers map[types.Address]struct{}) bool {
	aggregateVotingPower := big.NewInt(0)
	valsCount := 0

	for address := range signers {
		if votingPower := vs.votingPowerMap[address]; votingPower != nil {
			_ = aggregateVotingPower.Add(aggregateVotingPower, votingPower)

			valsCount++
		}
	}

	var quorumSize *big.Int
	if valsCount < 4 {
		quorumSize = vs.totalVotingPower
	} else {
		quorumSize = getQuorumSize(blockNumber, vs.totalVotingPower)
	}

	hasQuorum := aggregateVotingPower.Cmp(quorumSize) >= 0

	vs.logger.Debug("HasQuorum",
		"signers", len(signers),
		"signers voting power", aggregateVotingPower,
		"quorum size", quorumSize,
		"hasQuorum", hasQuorum)

	return hasQuorum
}

func (vs validatorSet) GetVotingPowers() map[string]*big.Int {
	result := make(map[string]*big.Int, vs.Len())

	for address, vp := range vs.votingPowerMap {
		result[types.AddressToString(address)] = vp
	}

	return result
}

func (vs validatorSet) Accounts() AccountSet {
	return vs.validators
}

func (vs validatorSet) Includes(address types.Address) bool {
	return vs.validators.ContainsAddress(address)
}

func (vs validatorSet) Len() int {
	return vs.validators.Len()
}

func (vs validatorSet) TotalVotingPower() big.Int {
	return *vs.totalVotingPower
}

// getQuorumSize calculates quorum size as 2/3 super-majority of provided total voting power
// Hydra modification: Qourum size is 61.4% (614/1000) of total voting power + 1
func getQuorumSize(blockNumber uint64, totalVotingPower *big.Int) *big.Int {
	quorum := new(big.Int)
	// In case for some reason voting power goes below too small value, we set quorum to total voting power
	if totalVotingPower.Cmp(big.NewInt(10)) == -1 {
		quorum = quorum.Add(quorum, totalVotingPower)
	} else {
		quorum = quorum.Mul(totalVotingPower, big.NewInt(614)).Div(quorum, big.NewInt(1000)).Add(quorum, big.NewInt(1))
	}

	return quorum
}
