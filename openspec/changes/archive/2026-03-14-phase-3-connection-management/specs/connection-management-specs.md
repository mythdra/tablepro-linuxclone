# Phase 3: Connection Management Specifications

## ADDED Requirements

### Capability: connection-models

The system SHALL provide data structures for database connections with SSH and SSL configuration.

#### Scenario: DatabaseConnection struct defined
- **WHEN** package loads
- **THEN** DatabaseConnection struct includes ID, Name, Type, Host, Port, Database, Username fields

#### Scenario: SSHTunnelConfig struct defined
- **WHEN** package loads
- **THEN** SSHTunnelConfig includes Enabled, Host, Port, Username, AuthMethod fields with password excluded from JSON

#### Scenario: SSLConfig struct defined
- **WHEN** package loads
- **THEN** SSLConfig includes Enabled, Mode, CACert, ClientCert fields with ClientKey excluded from JSON

#### Scenario: ConnectionSession struct defined
- **WHEN** package loads
- **THEN** ConnectionSession includes ConnectionID, Status, ActiveDB, LastPingAt fields

### Capability: connection-crud

The system SHALL provide CRUD operations for database connections with JSON persistence.

#### Scenario: Save connection
- **WHEN** user saves a connection
- **THEN** connection metadata is saved to ~/.config/tablepro/connections.json

#### Scenario: Load connections
- **WHEN** application starts
- **THEN** all connections are loaded from JSON file

#### Scenario: Delete connection
- **WHEN** user deletes a connection
- **THEN** connection is removed from JSON and password is deleted from keychain

#### Scenario: Duplicate connection
- **WHEN** user duplicates a connection
- **THEN** new connection has new UUID, name with " (Copy)" suffix, and copied credentials

#### Scenario: Update connection
- **WHEN** user updates connection details
- **THEN** changes are persisted to JSON file

#### Scenario: Validate connection
- **WHEN** connection is saved
- **THEN** Name and Type fields are required, invalid connections are rejected

### Capability: keychain-integration

The system SHALL store passwords securely in OS Keychain.

#### Scenario: Save password
- **WHEN** connection is saved with password
- **THEN** password is stored in OS Keychain with key "tablepro:password:{UUID}"

#### Scenario: Get password
- **WHEN** connection is established
- **THEN** password is retrieved from OS Keychain

#### Scenario: Delete password
- **WHEN** connection is deleted
- **THEN** password is removed from OS Keychain

#### Scenario: Handle keychain unavailable
- **WHEN** OS Keychain is not available (Linux server)
- **THEN** system logs warning and continues with in-memory fallback

### Capability: ssh-tunnel

The system SHALL establish SSH tunnels for database connections.

#### Scenario: SSH tunnel with password
- **WHEN** SSH is configured with password auth
- **THEN** tunnel establishes using ssh.Password() method

#### Scenario: SSH tunnel with key file
- **WHEN** SSH is configured with private key
- **THEN** tunnel establishes using ssh.PublicKeys(signer) method

#### Scenario: SSH tunnel with agent
- **WHEN** SSH agent is running and configured
- **THEN** tunnel uses keys from SSH_AUTH_SOCK

#### Scenario: Close SSH tunnel
- **WHEN** connection closes
- **THEN** SSH tunnel resources are released (listener and client closed)

#### Scenario: SSH health check
- **WHEN** tunnel is active
- **THEN** keepalive requests are sent every 30 seconds

### Capability: ssl-config

The system SHALL configure SSL/TLS for database connections.

#### Scenario: SSL mode disable
- **WHEN** SSL mode is "disable"
- **THEN** no SSL/TLS is used

#### Scenario: SSL mode require
- **WHEN** SSL mode is "require"
- **THEN** SSL is used without certificate verification

#### Scenario: SSL mode verify-ca
- **WHEN** SSL mode is "verify-ca"
- **THEN** CA certificate is verified

#### Scenario: SSL mode verify-full
- **WHEN** SSL mode is "verify-full"
- **THEN** CA certificate and server name are verified

#### Scenario: Client certificate authentication
- **WHEN** client cert and key are provided
- **THEN** mutual TLS is established

### Capability: url-parser

The system SHALL parse connection URLs including SSH URLs with dual @ symbols.

#### Scenario: Parse standard URL
- **WHEN** URL is postgres://host:port/db
- **THEN** host, port, and database are extracted correctly

#### Scenario: Parse SSH URL
- **WHEN** URL is postgres+ssh://sshuser@bastion:22/dbuser:pass@host:5432/db
- **THEN** SSH config and DB config are extracted separately

#### Scenario: Parse query parameters
- **WHEN** URL contains ?sslmode=require&statusColor=red
- **THEN** sslmode and statusColor are extracted

#### Scenario: Support all database schemes
- **WHEN** URL scheme is postgres, mysql, mongodb, or redis
- **THEN** scheme is recognized and mapped to DatabaseType

### Capability: deep-linking

The system SHALL handle tablepro:// deep links for opening connections.

#### Scenario: Register URL scheme
- **WHEN** app is installed
- **THEN** OS recognizes tablepro:// scheme

#### Scenario: Parse deep link
- **WHEN** tablepro://open?host=localhost&port=5432&db=mydb is received
- **THEN** parameters are extracted into connection config

#### Scenario: Queue links before ready
- **WHEN** deep link arrives before app startup complete
- **THEN** link is queued and processed after startup

#### Scenario: Auto-open connection
- **WHEN** deep link is processed
- **THEN** matching connection is opened or transient connection is created

### Capability: connection-testing

The system SHALL test database connectivity before saving.

#### Scenario: Test connection success
- **WHEN** user clicks Test Connection
- **THEN** system attempts connection and displays success message

#### Scenario: Test connection timeout
- **WHEN** connection takes more than 10 seconds
- **THEN** test is cancelled and timeout error is displayed

#### Scenario: Test connection with SSH
- **WHEN** SSH tunnel is enabled
- **THEN** tunnel is established before testing database connection

#### Scenario: Detailed error messages
- **WHEN** connection fails
- **THEN** error message includes actionable guidance (e.g., "check host and port")

### Capability: connection-ui

The system SHALL provide a React form for managing database connections.

#### Scenario: Connection form renders
- **WHEN** user opens connection dialog
- **THEN** form displays with tabs: General, SSH, SSL, Advanced

#### Scenario: Database type dropdown
- **WHEN** user selects database type
- **THEN** relevant fields are shown/hidden based on type

#### Scenario: Form validation
- **WHEN** required fields are empty
- **THEN** form cannot be submitted and shows validation errors

#### Scenario: Test connection button
- **WHEN** user clicks Test Connection
- **THEN** loading state is shown and result is displayed

#### Scenario: Save connection
- **WHEN** user clicks Save
- **THEN** connection is saved via ConnectionManager RPC call

#### Scenario: Status indicators
- **WHEN** connection status changes
- **THEN** UI shows colored dot (green=connected, red=error, grey=disconnected)
