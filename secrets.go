//go:generate mockgen -package=mock -destination=mock/secrets.mock.go github.com/libopenstorage/secrets Secrets
package secrets

import (
	"errors"
	"fmt"
)

var (
	// ErrNotSupported returned when implementation of specific function is not supported
	ErrNotSupported = errors.New("implementation not supported")
	// ErrNotAuthenticated returned when not authenticated with secrets endpoint
	ErrNotAuthenticated = errors.New("Not authenticated with the secrets endpoint")
	// ErrInvalidSecretId returned when no secret data is found associated with the id
	ErrInvalidSecretId = errors.New("No Secret Data found for Secret ID")
	// ErrEmptySecretData returned when no secret data is provided to store the secret
	ErrEmptySecretData = errors.New("Secret data cannot be empty")
	// ErrEmptySecretId returned when no secret Name/ID is provided to retrive secret data
	ErrEmptySecretId = errors.New("Secret Name/ID cannot be empty")
	// ErrSecretExists returned when a secret for the given secret id already exists
	ErrSecretExists = errors.New("Secret Id already exists")
	// ErrInvalidSecretData is returned when no secret data is found
	ErrInvalidSecretData = errors.New("Secret Data cannot be empty when CustomSecretData|PublicSecretData flag is set")
)

const (
	SecretPath = "/var/lib/osd/secrets/"
	// CustomSecretData is a constant used in the key context of the secrets APIs
	// It indicates that the secret provider should not generate secret but use the provided secret
	// in the API
	CustomSecretData = "custom_secret_data"
	// PublicSecretData is a constant used in the key context of Secret APIs
	// It indicates that the API is dealing with the public part of a secret instead
	// of the actual secret
	PublicSecretData = "public_secret_data"
	// OverwriteSecretDataInStore is a constant used in the key context of Secret APIs
	// It indicates whether the secret data stored in the persistent store can
	// be overwritten
	OverwriteSecretDataInStore = "overwrite_secret_data_in_store"
)

// Secrets interface implemented by backend Key Management Systems (KMS)
type Secrets interface {
	// String representation of the backend KMS
	String() string

	// GetSecret returns the secret data associated with the
	// supplied secretId. The secret data / plain text  can be used
	// by callers to encrypt their data. It is assumed that the plain text
	// data will be destroyed by the caller once used.
	GetSecret(
		secretId string,
		keyContext map[string]string,
	) (map[string]interface{}, error)

	// PutSecret will associate an secretId to its secret data
	// provided in the arguments and store it into the secret backend
	PutSecret(
		secretId string,
		plainText map[string]interface{},
		keyContext map[string]string,
	) error

	// DeleteSecret deletes the secret data associated with the
	// supplied secretId.
	DeleteSecret(
		secretId string,
		keyContext map[string]string,
	) error

	// Encrypt encrypts the supplied plain text data using the given key.
	// The API would fetch the plain text key, encrypt the data with it.
	// The plain text key will not be stored anywhere else and would be
	// deleted from memory.
	Encrypt(
		secretId string,
		plaintTextData string,
		keyContext map[string]string,
	) (string, error)

	// Decrypt decrypts the supplied encrypted  data using the given key.
	// The API would fetch the plain text key, decrypt the data with it.
	// The plain text key will not be stored anywhere else and would be
	// deleted from memory.
	Decrypt(
		secretId string,
		encryptedData string,
		keyContext map[string]string,
	) (string, error)

	// Reencrypt decrypts the data with the previous key and re-encrypts it
	// with the new key..
	Rencrypt(
		originalSecretId string,
		newSecretId string,
		originalKeyContext map[string]string,
		newKeyContext map[string]string,
		encryptedData string,
	) (string, error)

	// ListSecrets returns a list of known secretIDs
	ListSecrets() ([]string, error)
}

type BackendInit func(
	secretConfig map[string]interface{},
) (Secrets, error)

// ErrInvalidKeyContext is returned when secret data is provided to the secret APIs with an invalid
// key context.
type ErrInvalidKeyContext struct {
	Reason string
}

func (e *ErrInvalidKeyContext) Error() string {
	return fmt.Sprintf("invalid key context: %v", e.Reason)
}

// KeyContextChecks performs a series of checks on the keys and values
// passed through the key context map
func KeyContextChecks(
	keyContext map[string]string,
	secretData map[string]interface{},
) error {
	_, customData := keyContext[CustomSecretData]
	_, publicData := keyContext[PublicSecretData]

	if customData && publicData {
		return &ErrInvalidKeyContext{
			Reason: "both CustomSecretData and PublicSecretData flags cannot be set",
		}
	} else if !customData && !publicData && len(secretData) > 0 {
		return &ErrInvalidKeyContext{
			Reason: "secret data cannot be provided when none of CustomSecretData|PublicSecretData flag is not set",
		}
	} else if customData && len(secretData) == 0 {
		return &ErrInvalidKeyContext{
			Reason: "secret data needs to be provided when CustomSecretData flag is set",
		}
	} else if publicData && len(secretData) == 0 {
		return &ErrInvalidKeyContext{
			Reason: "secret data needs to be provided when PublicSecretData flag is set",
		}
	}
	return nil
}
