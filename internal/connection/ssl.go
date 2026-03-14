package connection

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

const (
	SSLModeDisable    = "disable"
	SSLModeRequire    = "require"
	SSLModeVerifyCA   = "verify-ca"
	SSLModeVerifyFull = "verify-full"
)

type SSLMode string

const (
	SSLModeDisableSSL    SSLMode = "disable"
	SSLModeRequireSSL    SSLMode = "require"
	SSLModeVerifyCASSL   SSLMode = "verify-ca"
	SSLModeVerifyFullSSL SSLMode = "verify-full"
)

var SSLModeMap = map[string]SSLMode{
	"disable":     SSLModeDisableSSL,
	"require":     SSLModeRequireSSL,
	"verify-ca":   SSLModeVerifyCASSL,
	"verify-full": SSLModeVerifyFullSSL,
	"verify":      SSLModeVerifyCASSL,
}

func parseAndValidateSSLMode(mode string) (SSLMode, error) {
	normalizedMode := normalizeSSLMode(mode)

	if sslMode, exists := SSLModeMap[normalizedMode]; exists {
		return sslMode, nil
	}

	return "", fmt.Errorf("invalid SSL mode: %s (valid modes: disable, require, verify-ca, verify-full)", mode)
}

func normalizeSSLMode(mode string) string {
	switch mode {
	case "disable", "disabled", "0", "false", "no":
		return "disable"
	case "require", "required", "1", "true", "yes":
		return "require"
	case "verify-ca", "verify ca", "verifyca":
		return "verify-ca"
	case "verify-full", "verify full", "verifyfull", "verify-identity":
		return "verify-full"
	default:
		return mode
	}
}

func BuildTLSConfig(cfg SSLConfig) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	mode, err := parseAndValidateSSLMode(cfg.Mode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSL mode: %w", err)
	}

	return buildTLSConfigForMode(mode, cfg)
}

func buildTLSConfigForMode(mode SSLMode, cfg SSLConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	switch mode {
	case SSLModeDisableSSL:
		return nil, nil

	case SSLModeRequireSSL:
		tlsConfig.InsecureSkipVerify = true
		return tlsConfig, nil

	case SSLModeVerifyCASSL:
		return buildVerifyCAConfig(cfg)

	case SSLModeVerifyFullSSL:
		return buildVerifyFullConfig(cfg)

	default:
		return nil, fmt.Errorf("unsupported SSL mode: %s", mode)
	}
}

func buildVerifyCAConfig(cfg SSLConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	if cfg.CACert != "" {
		caCertPool, err := loadCACertificate(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA certificate: %w", err)
		}
		tlsConfig.RootCAs = caCertPool
	}

	if cfg.ClientCert != "" {
		clientCert, err := loadClientCertificate(cfg.ClientCert)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

func buildVerifyFullConfig(cfg SSLConfig) (*tls.Config, error) {
	serverName := cfg.ServerName
	if serverName == "" {
		return nil, fmt.Errorf("ServerName is required for SSL mode 'verify-full'")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         serverName,
	}

	if cfg.CACert != "" {
		caCertPool, err := loadCACertificate(cfg.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA certificate: %w", err)
		}
		tlsConfig.RootCAs = caCertPool
	}

	if cfg.ClientCert != "" {
		clientCert, err := loadClientCertificate(cfg.ClientCert)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

func loadCertificate(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("certificate path is empty")
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("certificate file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to access certificate file: %w", err)
	}

	certBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return certBytes, nil
}

func loadCACertificate(path string) (*x509.CertPool, error) {
	certBytes, err := loadCertificate(path)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()

	if certPool.AppendCertsFromPEM(certBytes) {
		return certPool, nil
	}

	block, _ := pem.Decode(certBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse CA certificate: not a valid PEM format")
	}

	_, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	certPool.AppendCertsFromPEM(pem.EncodeToMemory(block))

	return certPool, nil
}

func loadClientCertificate(certPath string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, certPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load client certificate: %w", err)
	}

	return cert, nil
}

func ValidateSSLConfig(cfg SSLConfig) error {
	if !cfg.Enabled {
		return nil
	}

	_, err := parseAndValidateSSLMode(cfg.Mode)
	if err != nil {
		return err
	}

	if cfg.CACert != "" {
		if _, err := os.Stat(cfg.CACert); os.IsNotExist(err) {
			return fmt.Errorf("CA certificate file not found: %s", cfg.CACert)
		}
	}

	if cfg.ClientCert != "" {
		if _, err := os.Stat(cfg.ClientCert); os.IsNotExist(err) {
			return fmt.Errorf("client certificate file not found: %s", cfg.ClientCert)
		}
	}

	return nil
}

func GetSSLModeDescription(mode string) (string, error) {
	sslMode, err := parseAndValidateSSLMode(mode)
	if err != nil {
		return "", err
	}

	descriptions := map[SSLMode]string{
		SSLModeDisableSSL:    "SSL/TLS is disabled. Connection uses plain text.",
		SSLModeRequireSSL:    "SSL/TLS is required. Server certificate is not verified (insecure).",
		SSLModeVerifyCASSL:   "SSL/TLS is required. Server certificate is verified against CA certificate.",
		SSLModeVerifyFullSSL: "SSL/TLS is required. Server certificate is verified against CA and server name is validated.",
	}

	return descriptions[sslMode], nil
}

func SSLConfigWithDefaults(cfg SSLConfig) SSLConfig {
	result := cfg

	if result.Mode == "" {
		result.Mode = SSLModeRequire
	}

	return result
}
