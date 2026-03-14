package connection

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "TablePro"
	keyPrefix   = "tablepro:password:"
)

// KeychainError represents keychain-specific errors
type KeychainError struct {
	Op  string
	Err error
}

func (e *KeychainError) Error() string {
	return fmt.Sprintf("keychain %s failed: %v", e.Op, e.Err)
}

func (e *KeychainError) Unwrap() error {
	return e.Err
}

// ErrKeychainUnavailable is returned when the system keychain is unavailable
var ErrKeychainUnavailable = errors.New("keychain unavailable")

// ErrKeyNotFound is returned when a password is not found in the keychain
var ErrKeyNotFound = errors.New("password not found in keychain")

// buildKey creates the keychain key for a connection
func buildKey(connectionID uuid.UUID) string {
	return keyPrefix + connectionID.String()
}

// SavePassword stores a password in the OS keychain
func SavePassword(connectionID uuid.UUID, password string) error {
	if connectionID == uuid.Nil {
		return &KeychainError{Op: "save", Err: errors.New("connection ID cannot be nil")}
	}

	if password == "" {
		return &KeychainError{Op: "save", Err: errors.New("password cannot be empty")}
	}

	key := buildKey(connectionID)

	err := keyring.Set(serviceName, key, password)
	if err != nil {
		return handleKeychainError("save", err)
	}

	return nil
}

// GetPassword retrieves a password from the OS keychain
func GetPassword(connectionID uuid.UUID) (string, error) {
	if connectionID == uuid.Nil {
		return "", &KeychainError{Op: "get", Err: errors.New("connection ID cannot be nil")}
	}

	key := buildKey(connectionID)

	secret, err := keyring.Get(serviceName, key)
	if err != nil {
		return "", handleKeychainError("get", err)
	}

	return secret, nil
}

// DeletePassword removes a password from the OS keychain
func DeletePassword(connectionID uuid.UUID) error {
	if connectionID == uuid.Nil {
		return &KeychainError{Op: "delete", Err: errors.New("connection ID cannot be nil")}
	}

	key := buildKey(connectionID)

	err := keyring.Delete(serviceName, key)
	if err != nil {
		return handleKeychainError("delete", err)
	}

	return nil
}

// HasPassword checks if a password exists in the keychain
func HasPassword(connectionID uuid.UUID) bool {
	if connectionID == uuid.Nil {
		return false
	}

	key := buildKey(connectionID)
	_, err := keyring.Get(serviceName, key)
	return err == nil
}

// handleKeychainError converts keychain errors to appropriate application errors
func handleKeychainError(operation string, err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	switch {
	case strings.Contains(errMsg, "could not access"):
		return &KeychainError{
			Op:  operation,
			Err: ErrKeychainUnavailable,
		}
	case strings.Contains(errMsg, "not found"):
		return &KeychainError{
			Op:  operation,
			Err: ErrKeyNotFound,
		}
	case strings.Contains(errMsg, "secret service"):
		return &KeychainError{
			Op:  operation,
			Err: ErrKeychainUnavailable,
		}
	case strings.Contains(errMsg, "keychain"):
		return &KeychainError{
			Op:  operation,
			Err: ErrKeychainUnavailable,
		}
	case strings.Contains(errMsg, "credentials"):
		return &KeychainError{
			Op:  operation,
			Err: ErrKeychainUnavailable,
		}
	default:
		return &KeychainError{
			Op:  operation,
			Err: err,
		}
	}
}
