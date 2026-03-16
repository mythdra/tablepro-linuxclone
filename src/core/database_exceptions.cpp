#include "database_exceptions.h"

namespace tablepro {

DatabaseException::DatabaseException(const QString& message)
    : m_message(message)
{
}

const char* DatabaseException::what() const noexcept
{
    return m_message.toStdString().c_str();
}

ConnectionException::ConnectionException(const QString& message)
    : DatabaseException(message)
{
}

QueryExecutionException::QueryExecutionException(const QString& message)
    : DatabaseException(message)
{
}

TransactionException::TransactionException(const QString& message)
    : DatabaseException(message)
{
}

AuthenticationException::AuthenticationException(const QString& message)
    : DatabaseException(message)
{
}

} // namespace tablepro