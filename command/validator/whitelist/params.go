package whitelist

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/polygon-edge/command/helper"
	validatorHelper "github.com/0xPolygon/polygon-edge/command/validator/helper"
)

const (
	newValidatorAddressesFlag = "addresses"
)

var (
	errNoNewValidatorsProvided = errors.New("no new validators addresses provided")
)

type whitelistParams struct {
	accountDir            string
	accountConfig         string
	privateKey            string
	jsonRPC               string
	newValidatorAddresses []string
	txTimeout             time.Duration
	txPollFreq            time.Duration
}

func (ep *whitelistParams) validateFlags() error {
	if len(ep.newValidatorAddresses) == 0 {
		return errNoNewValidatorsProvided
	}

	if ep.privateKey == "" {
		return validatorHelper.ValidateSecretFlags(ep.accountDir, ep.accountConfig)
	}

	// validate jsonrpc address
	_, err := helper.ParseJSONRPCAddress(ep.jsonRPC)

	return err
}

type whitelistResult struct {
	NewValidatorAddresses []string `json:"newValidatorAddresses"`
}

func (wr whitelistResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[WHITELIST VALIDATORS]\n")

	vals := make([]string, len(wr.NewValidatorAddresses))
	for i, addr := range wr.NewValidatorAddresses {
		vals[i] = fmt.Sprintf("Validator address|%s", addr)
	}

	buffer.WriteString(helper.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}