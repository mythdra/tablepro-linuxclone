package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHTunnelConfig holds the configuration for an SSH tunnel.
type SSHTunnelConfig struct {
	// SSH server connection details
	Host     string // SSH server hostname or IP
	Port     int    // SSH server port (default 22)
	Username string // SSH username

	// Authentication method (only one should be set)
	Password       string // Password authentication
	PrivateKeyPath string // Path to private key file
	UseAgent       bool   // Use SSH agent for authentication

	// Key passphrase (optional, for encrypted private keys)
	KeyPassphrase string

	// Local port forwarding
	LocalPort  int    // Local port to listen on (0 = auto-assign)
	RemoteHost string // Remote host to forward to
	RemotePort int    // Remote port to forward to

	// Connection settings
	ConnectTimeout    time.Duration // Timeout for establishing SSH connection
	KeepAliveInterval time.Duration // Interval for keepalive requests
}

// SSHTunnel represents an active SSH tunnel with local port forwarding.
type SSHTunnel struct {
	Config    SSHTunnelConfig
	Client    *ssh.Client
	Listener  *net.TCPListener
	LocalPort int

	mu     sync.RWMutex
	closed bool
	done   chan struct{}
}

// NewSSHTunnel creates a new SSHTunnel with the given configuration.
func NewSSHTunnel(config SSHTunnelConfig) *SSHTunnel {
	return &SSHTunnel{
		Config: config,
		done:   make(chan struct{}),
	}
}

// Start establishes the SSH tunnel and starts listening for local connections.
func (t *SSHTunnel) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return fmt.Errorf("tunnel is already closed")
	}

	// Set defaults
	if t.Config.Port == 0 {
		t.Config.Port = 22
	}
	if t.Config.ConnectTimeout == 0 {
		t.Config.ConnectTimeout = 30 * time.Second
	}
	if t.Config.KeepAliveInterval == 0 {
		t.Config.KeepAliveInterval = 30 * time.Second
	}

	// Build SSH client config
	clientConfig, err := t.buildClientConfig()
	if err != nil {
		return fmt.Errorf("failed to build SSH client config: %w", err)
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", t.Config.Host, t.Config.Port)
	dialer := &net.Dialer{Timeout: t.Config.ConnectTimeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to dial SSH server %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, clientConfig)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	t.Client = ssh.NewClient(sshConn, chans, reqs)

	// Start keepalive goroutine
	go t.keepalive()

	// Start local listener for port forwarding
	localAddr := fmt.Sprintf("localhost:%d", t.Config.LocalPort)
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: t.Config.LocalPort})
	if err != nil {
		t.Client.Close()
		return fmt.Errorf("failed to listen on %s: %w", localAddr, err)
	}

	t.Listener = listener
	t.LocalPort = listener.Addr().(*net.TCPAddr).Port

	// Start forwarding connections
	go t.handleForwarding()

	return nil
}

// buildClientConfig builds the SSH client configuration with authentication.
func (t *SSHTunnel) buildClientConfig() (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	// SSH agent authentication - connect to SSH_AUTH_SOCK
	if t.Config.UseAgent {
		if authSock := os.Getenv("SSH_AUTH_SOCK"); authSock != "" {
			sshAgent, err := net.Dial("unix", authSock)
			if err == nil {
				agentClient := agent.NewClient(sshAgent)
				signers, err := agentClient.Signers()
				if err == nil && len(signers) > 0 {
					authMethods = append(authMethods, ssh.PublicKeys(signers...))
				}
			}
		}
	}

	// Private key file authentication
	if t.Config.PrivateKeyPath != "" {
		keySigner, err := t.parsePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(keySigner))
	}

	// Password authentication
	if t.Config.Password != "" {
		authMethods = append(authMethods, ssh.Password(t.Config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	return &ssh.ClientConfig{
		User:            t.Config.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         t.Config.ConnectTimeout,
	}, nil
}

// parsePrivateKey parses a private key file and returns an ssh.Signer.
func (t *SSHTunnel) parsePrivateKey() (ssh.Signer, error) {
	keyData, err := os.ReadFile(t.Config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	var signer ssh.Signer
	if t.Config.KeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(t.Config.KeyPassphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyData)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return signer, nil
}

// handleForwarding handles incoming connections and forwards them through the SSH tunnel.
func (t *SSHTunnel) handleForwarding() {
	for {
		select {
		case <-t.done:
			return
		default:
		}

		// Set accept deadline to allow checking done channel
		if t.Listener == nil {
			return
		}
		(*t.Listener).SetDeadline(time.Now().Add(1 * time.Second))

		conn, err := (*t.Listener).Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if netErr, ok := err.(net.Error); ok && !netErr.Temporary() {
				return
			}
			continue
		}

		go t.forwardConnection(conn)
	}
}

// forwardConnection forwards a single local connection to the remote target.
func (t *SSHTunnel) forwardConnection(localConn net.Conn) {
	defer localConn.Close()

	t.mu.RLock()
	client := t.Client
	t.mu.RUnlock()

	if client == nil {
		return
	}

	remoteAddr := fmt.Sprintf("%s:%d", t.Config.RemoteHost, t.Config.RemotePort)
	remoteConn, err := client.Dial("tcp", remoteAddr)
	if err != nil {
		return
	}
	defer remoteConn.Close()

	// Bidirectional copy
	done := make(chan struct{}, 2)

	go func() {
		io.Copy(localConn, remoteConn)
		localConn.Close()
		remoteConn.Close()
		done <- struct{}{}
	}()

	go func() {
		io.Copy(remoteConn, localConn)
		localConn.Close()
		remoteConn.Close()
		done <- struct{}{}
	}()

	<-done
}

// keepalive sends periodic keepalive requests to maintain the connection.
func (t *SSHTunnel) keepalive() {
	ticker := time.NewTicker(t.Config.KeepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.done:
			return
		case <-ticker.C:
			t.mu.RLock()
			client := t.Client
			t.mu.RUnlock()

			if client == nil {
				return
			}

			// Send a keepalive request
			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				// Connection might be dead, try to recover
				go t.reconnect()
				return
			}
		}
	}
}

// reconnect attempts to reconnect the SSH tunnel.
func (t *SSHTunnel) reconnect() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	// Close existing connection
	if t.Client != nil {
		t.Client.Close()
	}

	// Try to reconnect
	err := t.Start()
	if err != nil {
		// Log error but don't panic - let the health check handle reconnection
		fmt.Printf("SSH tunnel reconnection failed: %v\n", err)
	}
}

// Close terminates the SSH tunnel and releases all resources.
func (t *SSHTunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	close(t.done)

	// Close listener
	if t.Listener != nil {
		(*t.Listener).Close()
	}

	// Close SSH client
	if t.Client != nil {
		t.Client.Close()
	}

	return nil
}

// IsConnected returns whether the tunnel is currently connected.
func (t *SSHTunnel) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Client == nil || t.closed {
		return false
	}

	// Try a simple request to check if connection is alive
	_, _, err := t.Client.SendRequest("ping", true, nil)
	return err == nil
}

// LocalAddress returns the local address of the forwarded port.
func (t *SSHTunnel) LocalAddress() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Listener == nil {
		return ""
	}

	return (*t.Listener).Addr().String()
}
