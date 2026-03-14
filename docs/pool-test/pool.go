// Package session provides database session management with connection pooling.
package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConnectionPoolConfig holds configuration for a connection pool.
type ConnectionPoolConfig struct {
	MaxPoolSize int
	PoolTimeout time.Duration
}

// ConnectionPool manages a pool of reusable database connections.
type ConnectionPool struct {
	ConnectionID   uuid.UUID
	MaxPoolSize    int
	PoolTimeout    time.Duration
	mu             sync.RWMutex
	availableConns chan *sql.DB
	currentSize    int
	waitQueue      chan struct{}
	closed         bool
}

const DefaultPoolSize = 5
const DefaultPoolTimeout = 30 * time.Second

var ErrPoolExhausted = errors.New("connection pool exhausted: timeout waiting for available connection")

func NewConnectionPool(connID uuid.UUID, config *ConnectionPoolConfig) *ConnectionPool {
	maxSize := DefaultPoolSize
	timeout := DefaultPoolTimeout
	if config != nil {
		if config.MaxPoolSize > 0 {
			maxSize = config.MaxPoolSize
		}
		if config.PoolTimeout > 0 {
			timeout = config.PoolTimeout
		}
	}
	return &ConnectionPool{
		ConnectionID:   connID,
		MaxPoolSize:    maxSize,
		PoolTimeout:    timeout,
		availableConns: make(chan *sql.DB, maxSize),
		currentSize:    0,
		waitQueue:      make(chan struct{}, 1),
		closed:         false,
	}
}

func (p *ConnectionPool) addConnectionToPool(db *sql.DB) error {
	if p.closed {
		return fmt.Errorf("pool is closed")
	}
	select {
	case p.availableConns <- db:
		p.currentSize++
		return nil
	default:
		return fmt.Errorf("pool is full")
	}
}

func (p *ConnectionPool) createConnection(ctx context.Context, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, fmt.Errorf("pool is closed")
	}
	if p.currentSize >= p.MaxPoolSize {
		return nil, fmt.Errorf("pool at maximum capacity")
	}
	db, err := connFactory(ctx)
	if err != nil {
		return nil, fmt.Errorf("connection creation failed: %w", err)
	}
	if err := p.addConnectionToPool(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (p *ConnectionPool) GetConnection(ctx context.Context, connFactory func(context.Context) (*sql.DB, error)) (*sql.DB, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("pool is closed")
	}
	select {
	case db := <-p.availableConns:
		p.mu.Unlock()
		return db, nil
	default:
	}
	if p.currentSize < p.MaxPoolSize {
		p.mu.Unlock()
		return p.createConnection(ctx, connFactory)
	}
	p.mu.Unlock()
	timeoutCtx, cancel := context.WithTimeout(ctx, p.PoolTimeout)
	defer cancel()
	for {
		select {
		case <-timeoutCtx.Done():
			return nil, ErrPoolExhausted
		case <-p.waitQueue:
			p.mu.Lock()
			select {
			case db := <-p.availableConns:
				p.mu.Unlock()
				return db, nil
			default:
				p.mu.Unlock()
			}
		}
	}
}

func (p *ConnectionPool) ReturnConnection(db *sql.DB) {
	if db == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		db.Close()
		return
	}
	select {
	case p.availableConns <- db:
		select {
		case p.waitQueue <- struct{}{}:
		default:
		}
	default:
		db.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return
	}
	p.closed = true
	close(p.availableConns)
	for db := range p.availableConns {
		db.Close()
	}
	close(p.waitQueue)
}
