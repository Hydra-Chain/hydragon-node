package registration

import (
	"bytes"
	"fmt"

	"github.com/0xPolygon/polygon-edge/command/helper"
	sidechainHelper "github.com/0xPolygon/polygon-edge/command/sidechain"
	"github.com/0xPolygon/polygon-edge/helper/common"
)

const (
	stakeFlag = "stake"
)

type registerParams struct {
	accountDir         string
	accountConfig      string
	jsonRPC            string
	stake              string
	insecureLocalStore bool
}

func (rp *registerParams) validateFlags() error {
	if err := sidechainHelper.ValidateSecretFlags(rp.accountDir, rp.accountConfig); err != nil {
		return err
	}

	if _, err := helper.ParseJSONRPCAddress(rp.jsonRPC); err != nil {
		return fmt.Errorf("failed to parse json rpc address. Error: %w", err)
	}

	if rp.stake != "" {
		_, err := common.ParseUint256orHex(&rp.stake)
		if err != nil {
			return fmt.Errorf("provided stake '%s' isn't valid", rp.stake)
		}
	}

	return nil
}

type registerResult struct {
	validatorAddress string
	stakeResult      string
	amount           string
}

func (rr registerResult) GetOutput() string {
	var buffer bytes.Buffer

	var vals []string

	buffer.WriteString("\n[SUCCESSFUL REGISTRATION]\n")

	vals = make([]string, 0, 1)
	vals = append(vals, fmt.Sprintf("EVM Address|%s", rr.validatorAddress))

	buffer.WriteString(helper.FormatKV(vals))
	buffer.WriteString("\n")

	if rr.stakeResult != "" {
		buffer.WriteString("\n[SELF STAKE]\n")

		vals = make([]string, 0, 2)
		vals = append(vals, fmt.Sprintf("Staking Result|%s", rr.stakeResult))
		vals = append(vals, fmt.Sprintf("Amount Staked|%v", rr.amount))

		buffer.WriteString(helper.FormatKV(vals))
		buffer.WriteString("\n")
	}

	return buffer.String()
}
