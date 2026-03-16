#pragma once

#include <exception>
#include <QString>

namespace tablepro {

// Base exception class for database operations
class DatabaseException : public std::exception
{
public:
    explicit DatabaseException(const QString& message);
    const char* what() const noexcept override;

    QString message() const { return m_message; }

private:
    QString m_message;
};

// Specific exception classes
class ConnectionException : public DatabaseException
{
public:
    explicit ConnectionException(const QString& message);
};

class QueryExecutionException : public DatabaseException
{
public:
    explicit QueryExecutionException(const QString& message);
};

class TransactionException : public DatabaseException
{
public:
    explicit TransactionException(const QString& message);
};

class AuthenticationException : public DatabaseException
{
public:
    explicit AuthenticationException(const QString& message);
};

// Error codes for more granular error handling
enum class ErrorCode {
    Unknown = 0,
    ConnectionFailed,
    AuthenticationFailed,
    QueryExecutionFailed,
    Timeout,
    InvalidConfiguration,
    ConnectionLost,
    TransactionFailed
};

} // namespace tablepro