## ADDED Requirements

### Requirement: Session Event Emission
The system SHALL emit real-time events to the frontend when session state changes.

#### Scenario: Session created event
- **WHEN** a new session is successfully created
- **THEN** a `session:created` event is emitted with session UUID and connection details

#### Scenario: Session closed event
- **WHEN** a session is closed (by user or system)
- **THEN** a `session:closed` event is emitted with session UUID and close reason

#### Scenario: Session error event
- **WHEN** a session encounters an error (query failure, connection lost, etc.)
- **THEN** a `session:error` event is emitted with session UUID and error message

#### Scenario: Session reconnecting event
- **WHEN** the system begins attempting to reconnect a session
- **THEN** a `session:reconnecting` event is emitted with session UUID and retry count

### Requirement: Frontend Event Subscription
The frontend SHALL subscribe to session events to display real-time status.

#### Scenario: Subscribe on app startup
- **WHEN** the React app initializes
- **THEN** event listeners are registered for all session events using `EventsOn`

#### Scenario: Unsubscribe on app shutdown
- **WHEN** the app is unmounting
- **THEN** event listeners are removed using `EventsOff` to prevent memory leaks

#### Scenario: Display session status in UI
- **WHEN** a `session:created` or `session:closed` event is received
- **THEN** the sidebar connection status indicator updates to show connected/disconnected state

#### Scenario: Show error toast on session error
- **WHEN** a `session:error` event is received
- **THEN** a toast notification displays the error message to the user

#### Scenario: Show reconnecting indicator
- **WHEN** a `session:reconnecting` event is received
- **THEN** the UI displays a "Reconnecting..." indicator with retry count

### Requirement: Event Payload Structure
Session events SHALL include standardized payload data for frontend consumption.

#### Scenario: session:created payload
- **WHEN** `session:created` is emitted
- **THEN** payload includes: `{ sessionId: string, connectionName: string, databaseType: string, timestamp: number }`

#### Scenario: session:closed payload
- **WHEN** `session:closed` is emitted
- **THEN** payload includes: `{ sessionId: string, reason: string, timestamp: number }`

#### Scenario: session:error payload
- **WHEN** `session:error` is emitted
- **THEN** payload includes: `{ sessionId: string, error: string, isRecoverable: boolean, timestamp: number }`

#### Scenario: session:reconnecting payload
- **WHEN** `session:reconnecting` is emitted
- **THEN** payload includes: `{ sessionId: string, retryCount: number, nextRetryIn: number, timestamp: number }`
