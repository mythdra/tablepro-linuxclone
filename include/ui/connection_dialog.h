#pragma once

#include <QDialog>
#include <QLineEdit>
#include <QComboBox>
#include <QSpinBox>
#include <QCheckBox>
#include <QPushButton>
#include <QLabel>
#include <QDialogButtonBox>
#include <QGroupBox>
#include "core/connection_config.h"
#include "core/database_types.h"

namespace tablepro {

class DatabaseDriver;

/**
 * Dialog for configuring and testing database connections.
 */
class ConnectionDialog : public QDialog
{
    Q_OBJECT

public:
    explicit ConnectionDialog(QWidget* parent = nullptr);
    ~ConnectionDialog() override;

    ConnectionConfig getConnectionConfig() const;
    void setConnectionConfig(const ConnectionConfig& config);

    void setDriver(DatabaseDriver* driver);

    // Recent connections
    void setRecentConnections(const QStringList& connections);
    QStringList recentConnections() const;

signals:
    void connectionCreated(const ConnectionConfig& config);
    void connectionTested(const ConnectionConfig& config, bool success, const QString& message);

private slots:
    void testConnection();
    void saveConnection();
    void loadConnection();
    void onDatabaseTypeChanged(int index);
    void updateConnectionName();
    void onRecentConnectionSelected(int index);

private:
    void setupUI();
    void setupConnections();
    void loadSettings();
    void saveSettings();
    QString buildConnectionName() const;
    void updateUIForDatabaseType(DatabaseType type);

    // Connection settings
    QComboBox* m_typeCombo;
    QLineEdit* m_nameEdit;
    QLineEdit* m_hostEdit;
    QSpinBox* m_portSpinBox;
    QLineEdit* m_databaseEdit;
    QLineEdit* m_usernameEdit;
    QLineEdit* m_passwordEdit;

    // SSL settings
    QGroupBox* m_sslGroup;
    QCheckBox* m_useSslCheck;
    QLineEdit* m_sslCertEdit;
    QPushButton* m_sslCertBrowseBtn;

    // SSH settings
    QGroupBox* m_sshGroup;
    QCheckBox* m_useSshCheck;
    QLineEdit* m_sshHostEdit;
    QSpinBox* m_sshPortSpinBox;
    QLineEdit* m_sshUsernameEdit;
    QLineEdit* m_sshKeyEdit;
    QPushButton* m_sshKeyBrowseBtn;

    // Advanced settings
    QSpinBox* m_timeoutSpinBox;

    // Recent connections
    QComboBox* m_recentCombo;

    // Dialog buttons
    QPushButton* m_testBtn;
    QPushButton* m_saveBtn;
    QPushButton* m_connectBtn;
    QDialogButtonBox* m_buttonBox;

    // Status
    QLabel* m_statusLabel;

    // Data
    DatabaseDriver* m_driver = nullptr;
    QStringList m_recentConnectionList;
};

} // namespace tablepro