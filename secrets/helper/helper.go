package helper

import (
	"errors"
	"fmt"

	bls "github.com/0xPolygon/polygon-edge/consensus/polybft/signer"
	"github.com/0xPolygon/polygon-edge/consensus/polybft/wallet"
	"github.com/hashicorp/go-hclog"
	libp2pCrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/0xPolygon/polygon-edge/crypto"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/network"
	"github.com/0xPolygon/polygon-edge/secrets"
	"github.com/0xPolygon/polygon-edge/secrets/awsssm"
	"github.com/0xPolygon/polygon-edge/secrets/encryptedlocal"
	"github.com/0xPolygon/polygon-edge/secrets/gcpssm"
	"github.com/0xPolygon/polygon-edge/secrets/hashicorpvault"
	"github.com/0xPolygon/polygon-edge/secrets/local"
	"github.com/0xPolygon/polygon-edge/types"
)

// SetupEncryptedLocalSecretsManager is a helper method for encrypted local secrets manager setup
func SetupEncryptedLocalSecretsManager(dataDir string) (secrets.SecretsManager, error) {
	return encryptedlocal.SecretsManagerFactory(
		nil, // Local secrets manager doesn't require a config
		&secrets.SecretsManagerParams{
			Logger: hclog.New(&hclog.LoggerOptions{
				Name:  "secrets",
				Level: hclog.Info,
			}),
			Extra: map[string]interface{}{
				secrets.Path: dataDir,
			},
		},
	)
}

// SetupLocalSecretsManager is a helper method for boilerplate local secrets manager setup
func SetupLocalSecretsManager(dataDir string) (secrets.SecretsManager, error) {
	return local.SecretsManagerFactory(
		nil, // Local secrets manager doesn't require a config
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
			Extra: map[string]interface{}{
				secrets.Path: dataDir,
			},
		},
	)
}

// setupHashicorpVault is a helper method for boilerplate hashicorp vault secrets manager setup
func setupHashicorpVault(
	secretsConfig *secrets.SecretsManagerConfig,
) (secrets.SecretsManager, error) {
	return hashicorpvault.SecretsManagerFactory(
		secretsConfig,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
		},
	)
}

// setupAWSSSM is a helper method for boilerplate aws ssm secrets manager setup
func setupAWSSSM(
	secretsConfig *secrets.SecretsManagerConfig,
) (secrets.SecretsManager, error) {
	return awsssm.SecretsManagerFactory(
		secretsConfig,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
		},
	)
}

// setupGCPSSM is a helper method for boilerplate Google Cloud Computing secrets manager setup
func setupGCPSSM(
	secretsConfig *secrets.SecretsManagerConfig,
) (secrets.SecretsManager, error) {
	return gcpssm.SecretsManagerFactory(
		secretsConfig,
		&secrets.SecretsManagerParams{
			Logger: hclog.NewNullLogger(),
		},
	)
}

// InitECDSAValidatorKey creates new ECDSA key and set as a validator key
func InitECDSAValidatorKey(secretsManager secrets.SecretsManager) (types.Address, error) {
	if secretsManager.HasSecret(secrets.ValidatorKey) {
		return types.ZeroAddress, fmt.Errorf(
			`secrets "%s" has been already initialized`,
			secrets.ValidatorKey,
		)
	}

	validatorKey, validatorKeyEncoded, err := crypto.GenerateAndEncodeECDSAPrivateKey()
	if err != nil {
		return types.ZeroAddress, err
	}

	address := crypto.PubKeyToAddress(&validatorKey.PublicKey)

	// Write the validator private key to the secrets manager storage
	if setErr := secretsManager.SetSecret(
		secrets.ValidatorKey,
		validatorKeyEncoded,
	); setErr != nil {
		return types.ZeroAddress, setErr
	}

	return address, nil
}

func InitBLSValidatorKey(secretsManager secrets.SecretsManager) ([]byte, error) {
	if secretsManager.HasSecret(secrets.ValidatorBLSKey) {
		return nil, fmt.Errorf(`secrets "%s" has been already initialized`, secrets.ValidatorBLSKey)
	}

	blsSecretKey, blsSecretKeyEncoded, err := crypto.GenerateAndEncodeBLSSecretKey()
	if err != nil {
		return nil, err
	}

	// Write the validator private key to the secrets manager storage
	if setErr := secretsManager.SetSecret(
		secrets.ValidatorBLSKey,
		blsSecretKeyEncoded,
	); setErr != nil {
		return nil, setErr
	}

	pubkeyBytes, err := crypto.BLSSecretKeyToPubkeyBytes(blsSecretKey)
	if err != nil {
		return nil, err
	}

	return pubkeyBytes, nil
}

func InitNetworkingPrivateKey(
	secretsManager secrets.SecretsManager,
	predefinedSecret []byte,
) (libp2pCrypto.PrivKey, error) {
	if secretsManager.HasSecret(secrets.NetworkKey) {
		return nil, fmt.Errorf(`secrets "%s" has been already initialized`, secrets.NetworkKey)
	}

	var (
		libp2pKey        libp2pCrypto.PrivKey
		libp2pKeyEncoded []byte
		err              error
	)

	if predefinedSecret != nil && len(predefinedSecret) > 0 {
		libp2pKey, err = network.ParseLibp2pKey(predefinedSecret)
		if err != nil {
			return nil, fmt.Errorf("invalid private network key provided: %w", err)
		}

		libp2pKeyEncoded = predefinedSecret
	} else {
		// Generate the libp2p private key
		libp2pKey, libp2pKeyEncoded, err = network.GenerateAndEncodeLibp2pKey()
		if err != nil {
			return nil, err
		}
	}

	// Write the networking private key to the secrets manager storage
	if setErr := secretsManager.SetSecret(
		secrets.NetworkKey,
		libp2pKeyEncoded,
	); setErr != nil {
		return nil, setErr
	}

	return libp2pKey, err
}

// LoadValidatorAddress loads ECDSA key by SecretsManager and returns validator address
func LoadValidatorAddress(secretsManager secrets.SecretsManager) (types.Address, error) {
	if !secretsManager.HasSecret(secrets.ValidatorKey) {
		return types.ZeroAddress, nil
	}

	encodedKey, err := secretsManager.GetSecret(secrets.ValidatorKey)
	if err != nil {
		return types.ZeroAddress, err
	}

	privateKey, err := crypto.BytesToECDSAPrivateKey(encodedKey)
	if err != nil {
		return types.ZeroAddress, err
	}

	return crypto.PubKeyToAddress(&privateKey.PublicKey), nil
}

// LoadBLSPublicKey loads BLS key by SecretsManager and returns BLS Public Key
func LoadBLSPublicKey(secretsManager secrets.SecretsManager) (string, error) {
	if !secretsManager.HasSecret(secrets.ValidatorBLSKey) {
		return "", nil
	}

	encodedKey, err := secretsManager.GetSecret(secrets.ValidatorBLSKey)
	if err != nil {
		return "", err
	}

	secretKey, err := crypto.BytesToBLSSecretKey(encodedKey)
	if err != nil {
		return "", err
	}

	pubkeyBytes, err := crypto.BLSSecretKeyToPubkeyBytes(secretKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToHex(pubkeyBytes), nil
}

// LoadNodeID loads Libp2p key by SecretsManager and returns Node ID
func LoadNodeID(secretsManager secrets.SecretsManager) (string, error) {
	if !secretsManager.HasSecret(secrets.NetworkKey) {
		return "", nil
	}

	encodedKey, err := secretsManager.GetSecret(secrets.NetworkKey)
	if err != nil {
		return "", err
	}

	parsedKey, err := network.ParseLibp2pKey(encodedKey)
	if err != nil {
		return "", err
	}

	nodeID, err := peer.IDFromPrivateKey(parsedKey)
	if err != nil {
		return "", err
	}

	return nodeID.String(), nil
}

// InitCloudSecretsManager returns the cloud secrets manager from the provided config
func InitCloudSecretsManager(
	secretsConfig *secrets.SecretsManagerConfig,
) (secrets.SecretsManager, error) {
	var secretsManager secrets.SecretsManager

	switch secretsConfig.Type {
	case secrets.HashicorpVault:
		vault, err := setupHashicorpVault(secretsConfig)
		if err != nil {
			return secretsManager, err
		}

		secretsManager = vault
	case secrets.AWSSSM:
		AWSSSM, err := setupAWSSSM(secretsConfig)
		if err != nil {
			return secretsManager, err
		}

		secretsManager = AWSSSM
	case secrets.GCPSSM:
		GCPSSM, err := setupGCPSSM(secretsConfig)
		if err != nil {
			return secretsManager, err
		}

		secretsManager = GCPSSM
	default:
		return secretsManager, errors.New("unsupported secrets manager")
	}

	return secretsManager, nil
}

// H_MODIFY: add new func to init BLS signature to be used in the initial setup
func InitValidatorBLSSignature(
	secretsManager secrets.SecretsManager, account *wallet.Account, chainID int64) ([]byte, error) {
	if secretsManager.HasSecret(secrets.ValidatorBLSSignature) {
		return nil, fmt.Errorf(
			`secrets "%s" has been already initialized`,
			secrets.ValidatorBLSSignature,
		)
	}

	// Generate the signature
	s, err := bls.MakeKOSKSignature(
		account.Bls,
		types.Address(account.Ecdsa.Address()),
		chainID,
		bls.DomainHydraChain,
	)
	if err != nil {
		return nil, err
	}

	sb, err := s.Marshal()
	if err != nil {
		return nil, err
	}

	// Write the signature to the secrets manager storage
	if err := secretsManager.SetSecret(
		secrets.ValidatorBLSSignature,
		[]byte(hex.EncodeToString(sb)),
	); err != nil {
		return nil, err
	}

	return sb, nil
}

// LoadBLSSignature loads BLS Signature from SecretsManager and returns it
func LoadBLSSignature(secretsManager secrets.SecretsManager) (string, error) {
	if !secretsManager.HasSecret(secrets.ValidatorBLSSignature) {
		return "", nil
	}

	s, err := secretsManager.GetSecret(secrets.ValidatorBLSSignature)

	return string(s), err
}

// HydraGon Functions
// LoadEncodedSecret retrieves the specified secret by its name from SecretsManager and returns the
// encoded secret value
func LoadEncodedSecret(secretsManager secrets.SecretsManager, name string) ([]byte, error) {
	if !secretsManager.HasSecret(name) {
		return nil, fmt.Errorf("%s not exist", name)
	}

	encodedValue, err := secretsManager.GetSecret(name)
	if err != nil {
		return nil, err
	}

	return encodedValue, nil
}

// GetValidatorAddress loads ECDSA key by SecretsManager and returns validator address
func GetValidatorAddress(secretsManager secrets.SecretsManager) (string, error) {
	ecdsaKey, err := wallet.GetEcdsaFromSecret(secretsManager)
	if err != nil {
		return "", err
	}

	return ecdsaKey.Address().String(), nil
}

// GetBLSPublicKey loads BLS key by SecretsManager and returns BLS Public Key
func GetBLSPublicKey(secretsManager secrets.SecretsManager) (string, error) {
	bls, err := wallet.GetBlsFromSecret(secretsManager)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bls.PublicKey().Marshal()), nil
}
