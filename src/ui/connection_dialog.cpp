#include "connection_dialog.h"
#include "core/database_driver.h"
#include "drivers/postgres_driver.h"

#include <QFormLayout>
#include <QVBoxLayout>
#include <QHBoxLayout>
#include <QGridLayout>
#include <QFileDialog>
#include <QSettings>
#include <QMessageBox>
#include <QApplication>
#include <QDebug>

namespace tablepro {

ConnectionDialog::ConnectionDialog(QWidget* parent)
    : QDialog(parent)
{
    setupUI();
    setupConnections();
    loadSettings();

    // Set default port based on database type
    onDatabaseTypeChanged(0);
}

ConnectionDialog::~ConnectionDialog()
{
    saveSettings();
}

void ConnectionDialog::setupUI()
{
    setWindowTitle(tr("New Database Connection"));
    setMinimumWidth(500);

    auto* mainLayout = new QVBoxLayout(this);

    // Recent connections
    auto* recentGroup = new QGroupBox(tr("Recent Connections"), this);
    auto* recentLayout = new QHBoxLayout(recentGroup);
    m_recentCombo = new QComboBox(this);
    m_recentCombo->setMinimumWidth(300);
    recentLayout->addWidget(m_recentCombo);
    mainLayout->addWidget(recentGroup);

    // Connection settings
    auto* connGroup = new QGroupBox(tr("Connection Settings"), this);
    auto* connLayout = new QFormLayout(connGroup);

    // Connection name
    m_nameEdit = new QLineEdit(this);
    m_nameEdit->setPlaceholderText(tr("Auto-generated from connection details"));
    connLayout->addRow(tr("Name:"), m_nameEdit);

    // Database type
    m_typeCombo = new QComboBox(this);
    m_typeCombo->addItem(tr("PostgreSQL"), static_cast<int>(DatabaseType::PostgreSQL));
    m_typeCombo->addItem(tr("MySQL"), static_cast<int>(DatabaseType::MySQL));
    m_typeCombo->addItem(tr("SQLite"), static_cast<int>(DatabaseType::SQLite));
    m_typeCombo->addItem(tr("DuckDB"), static_cast<int>(DatabaseType::DuckDB));
    connLayout->addRow(tr("Type:"), m_typeCombo);

    // Host
    m_hostEdit = new QLineEdit(this);
    m_hostEdit->setPlaceholderText("localhost");
    connLayout->addRow(tr("Host:"), m_hostEdit);

    // Port
    m_portSpinBox = new QSpinBox(this);
    m_portSpinBox->setRange(1, 65535);
    m_portSpinBox->setValue(5432);
    connLayout->addRow(tr("Port:"), m_portSpinBox);

    // Database
    m_databaseEdit = new QLineEdit(this);
    connLayout->addRow(tr("Database:"), m_databaseEdit);

    // Username
    m_usernameEdit = new QLineEdit(this);
    connLayout->addRow(tr("Username:"), m_usernameEdit);

    // Password
    m_passwordEdit = new QLineEdit(this);
    m_passwordEdit->setEchoMode(QLineEdit::Password);
    connLayout->addRow(tr("Password:"), m_passwordEdit);

    mainLayout->addWidget(connGroup);

    // SSL settings
    m_sslGroup = new QGroupBox(tr("SSL Settings"), this);
    auto* sslLayout = new QGridLayout(m_sslGroup);

    m_useSslCheck = new QCheckBox(tr("Use SSL"), this);
    sslLayout->addWidget(m_useSslCheck, 0, 0, 1, 2);

    m_sslCertEdit = new QLineEdit(this);
    m_sslCertBrowseBtn = new QPushButton(tr("Browse..."), this);
    sslLayout->addWidget(new QLabel(tr("Certificate:")), 1, 0);
    sslLayout->addWidget(m_sslCertEdit, 1, 1);
    sslLayout->addWidget(m_sslCertBrowseBtn, 1, 2);

    mainLayout->addWidget(m_sslGroup);

    // SSH settings
    m_sshGroup = new QGroupBox(tr("SSH Tunnel"), this);
    auto* sshLayout = new QGridLayout(m_sshGroup);

    m_useSshCheck = new QCheckBox(tr("Use SSH Tunnel"), this);
    sshLayout->addWidget(m_useSshCheck, 0, 0, 1, 4);

    sshLayout->addWidget(new QLabel(tr("Host:")), 1, 0);
    m_sshHostEdit = new QLineEdit(this);
    sshLayout->addWidget(m_sshHostEdit, 1, 1);

    sshLayout->addWidget(new QLabel(tr("Port:")), 1, 2);
    m_sshPortSpinBox = new QSpinBox(this);
    m_sshPortSpinBox->setRange(1, 65535);
    m_sshPortSpinBox->setValue(22);
    sshLayout->addWidget(m_sshPortSpinBox, 1, 3);

    sshLayout->addWidget(new QLabel(tr("Username:")), 2, 0);
    m_sshUsernameEdit = new QLineEdit(this);
    sshLayout->addWidget(m_sshUsernameEdit, 2, 1);

    sshLayout->addWidget(new QLabel(tr("Key:")), 3, 0);
    m_sshKeyEdit = new QLineEdit(this);
    sshLayout->addWidget(m_sshKeyEdit, 3, 1);
    m_sshKeyBrowseBtn = new QPushButton(tr("Browse..."), this);
    sshLayout->addWidget(m_sshKeyBrowseBtn, 3, 2);

    mainLayout->addWidget(m_sshGroup);

    // Advanced settings
    auto* advGroup = new QGroupBox(tr("Advanced"), this);
    auto* advLayout = new QFormLayout(advGroup);

    m_timeoutSpinBox = new QSpinBox(this);
    m_timeoutSpinBox->setRange(1, 300);
    m_timeoutSpinBox->setValue(30);
    m_timeoutSpinBox->setSuffix(tr(" seconds"));
    advLayout->addRow(tr("Connection Timeout:"), m_timeoutSpinBox);

    mainLayout->addWidget(advGroup);

    // Status label
    m_statusLabel = new QLabel(this);
    m_statusLabel->setWordWrap(true);
    mainLayout->addWidget(m_statusLabel);

    // Dialog buttons
    m_buttonBox = new QDialogButtonBox(this);
    m_testBtn = m_buttonBox->addButton(tr("Test"), QDialogButtonBox::ActionRole);
    m_saveBtn = m_buttonBox->addButton(tr("Save"), QDialogButtonBox::ActionRole);
    m_connectBtn = m_buttonBox->addButton(tr("Connect"), QDialogButtonBox::AcceptRole);
    m_buttonBox->addButton(QDialogButtonBox::Cancel);

    mainLayout->addWidget(m_buttonBox);
}

void ConnectionDialog::setupConnections()
{
    connect(m_typeCombo, QOverload<int>::of(&QComboBox::currentIndexChanged),
            this, &ConnectionDialog::onDatabaseTypeChanged);

    connect(m_hostEdit, &QLineEdit::textChanged, this, &ConnectionDialog::updateConnectionName);
    connect(m_databaseEdit, &QLineEdit::textChanged, this, &ConnectionDialog::updateConnectionName);
    connect(m_usernameEdit, &QLineEdit::textChanged, this, &ConnectionDialog::updateConnectionName);

    connect(m_recentCombo, QOverload<int>::of(&QComboBox::currentIndexChanged),
            this, &ConnectionDialog::onRecentConnectionSelected);

    connect(m_sslCertBrowseBtn, &QPushButton::clicked, this, [this]() {
        QString file = QFileDialog::getOpenFileName(this, tr("Select SSL Certificate"));
        if (!file.isEmpty()) {
            m_sslCertEdit->setText(file);
        }
    });

    connect(m_sshKeyBrowseBtn, &QPushButton::clicked, this, [this]() {
        QString file = QFileDialog::getOpenFileName(this, tr("Select SSH Private Key"));
        if (!file.isEmpty()) {
            m_sshKeyEdit->setText(file);
        }
    });

    connect(m_testBtn, &QPushButton::clicked, this, &ConnectionDialog::testConnection);
    connect(m_saveBtn, &QPushButton::clicked, this, &ConnectionDialog::saveConnection);
    connect(m_buttonBox, &QDialogButtonBox::accepted, this, &QDialog::accept);
    connect(m_buttonBox, &QDialogButtonBox::rejected, this, &QDialog::reject);
}

ConnectionConfig ConnectionDialog::getConnectionConfig() const
{
    ConnectionConfig config;

    config.type = static_cast<DatabaseType>(m_typeCombo->currentData().toInt());
    config.host = m_hostEdit->text().trimmed();
    config.port = m_portSpinBox->value();
    config.database = m_databaseEdit->text().trimmed();
    config.username = m_usernameEdit->text().trimmed();
    config.password = m_passwordEdit->text();
    config.useSsl = m_useSslCheck->isChecked();
    config.sslCertPath = m_sslCertEdit->text().trimmed();
    config.sshHost = m_sshHostEdit->text().trimmed();
    config.sshPort = m_sshPortSpinBox->value();
    config.sshUsername = m_sshUsernameEdit->text().trimmed();
    config.sshPrivateKeyPath = m_sshKeyEdit->text().trimmed();
    config.timeout = m_timeoutSpinBox->value();

    return config;
}

void ConnectionDialog::setConnectionConfig(const ConnectionConfig& config)
{
    int typeIndex = m_typeCombo->findData(static_cast<int>(config.type));
    if (typeIndex >= 0) {
        m_typeCombo->setCurrentIndex(typeIndex);
    }

    m_hostEdit->setText(config.host);
    m_portSpinBox->setValue(config.port);
    m_databaseEdit->setText(config.database);
    m_usernameEdit->setText(config.username);
    m_passwordEdit->setText(config.password);
    m_useSslCheck->setChecked(config.useSsl);
    m_sslCertEdit->setText(config.sslCertPath);
    m_sshHostEdit->setText(config.sshHost);
    m_sshPortSpinBox->setValue(config.sshPort);
    m_sshUsernameEdit->setText(config.sshUsername);
    m_sshKeyEdit->setText(config.sshPrivateKeyPath);
    m_timeoutSpinBox->setValue(config.timeout);

    updateConnectionName();
}

void ConnectionDialog::setDriver(DatabaseDriver* driver)
{
    m_driver = driver;
}

void ConnectionDialog::setRecentConnections(const QStringList& connections)
{
    m_recentConnectionList = connections;
    m_recentCombo->clear();
    m_recentCombo->addItem(tr("-- Select a recent connection --"));
    m_recentCombo->addItems(connections);
}

QStringList ConnectionDialog::recentConnections() const
{
    return m_recentConnectionList;
}

void ConnectionDialog::testConnection()
{
    ConnectionConfig config = getConnectionConfig();

    if (!config.isValid()) {
        m_statusLabel->setText(tr("Please fill in all required fields"));
        m_statusLabel->setStyleSheet("color: red;");
        return;
    }

    // Create a temporary driver based on database type
    DatabaseDriver* testDriver = nullptr;

    switch (config.type) {
    case DatabaseType::PostgreSQL:
        testDriver = new PostgresDriver();
        break;
    case DatabaseType::MySQL:
        m_statusLabel->setText(tr("MySQL driver not yet implemented"));
        m_statusLabel->setStyleSheet("color: orange;");
        return;
    case DatabaseType::SQLite:
        m_statusLabel->setText(tr("SQLite driver not yet implemented"));
        m_statusLabel->setStyleSheet("color: orange;");
        return;
    default:
        m_statusLabel->setText(tr("Unsupported database type"));
        m_statusLabel->setStyleSheet("color: red;");
        return;
    }

    m_statusLabel->setText(tr("Testing connection..."));
    m_statusLabel->setStyleSheet("color: blue;");
    QApplication::processEvents();

    bool success = testDriver->connect(config);

    if (success) {
        m_statusLabel->setText(tr("✓ Connection successful!"));
        m_statusLabel->setStyleSheet("color: green; font-weight: bold;");
        testDriver->disconnect();
        emit connectionTested(config, true, tr("Connection successful"));
    } else {
        m_statusLabel->setText(tr("✗ Connection failed: %1").arg(testDriver->lastError()));
        m_statusLabel->setStyleSheet("color: red;");
        emit connectionTested(config, false, testDriver->lastError());
    }

    delete testDriver;
}

void ConnectionDialog::saveConnection()
{
    ConnectionConfig config = getConnectionConfig();
    QString name = m_nameEdit->text().trimmed();

    if (name.isEmpty()) {
        name = buildConnectionName();
    }

    QSettings settings;
    settings.beginGroup("Connections");
    settings.setValue(name + "/type", static_cast<int>(config.type));
    settings.setValue(name + "/host", config.host);
    settings.setValue(name + "/port", config.port);
    settings.setValue(name + "/database", config.database);
    settings.setValue(name + "/username", config.username);
    settings.setValue(name + "/useSsl", config.useSsl);
    settings.setValue(name + "/sslCertPath", config.sslCertPath);
    settings.setValue(name + "/timeout", config.timeout);
    settings.endGroup();

    m_statusLabel->setText(tr("Connection '%1' saved").arg(name));
    emit connectionCreated(config);
}

void ConnectionDialog::loadConnection()
{
    // Load from recent connections
}

void ConnectionDialog::onDatabaseTypeChanged(int index)
{
    DatabaseType type = static_cast<DatabaseType>(m_typeCombo->itemData(index).toInt());
    updateUIForDatabaseType(type);
    updateConnectionName();
}

void ConnectionDialog::updateConnectionName()
{
    if (m_nameEdit->text().isEmpty() || m_nameEdit->text() == buildConnectionName()) {
        m_nameEdit->setText(buildConnectionName());
    }
}

void ConnectionDialog::onRecentConnectionSelected(int index)
{
    if (index <= 0) return;  // Skip placeholder

    QString name = m_recentCombo->itemText(index);
    QSettings settings;
    settings.beginGroup("Connections");

    if (settings.contains(name + "/host")) {
        ConnectionConfig config;
        config.type = static_cast<DatabaseType>(settings.value(name + "/type").toInt());
        config.host = settings.value(name + "/host").toString();
        config.port = settings.value(name + "/port").toInt();
        config.database = settings.value(name + "/database").toString();
        config.username = settings.value(name + "/username").toString();
        config.useSsl = settings.value(name + "/useSsl").toBool();
        config.sslCertPath = settings.value(name + "/sslCertPath").toString();
        config.timeout = settings.value(name + "/timeout", 30).toInt();

        setConnectionConfig(config);
        m_nameEdit->setText(name);
    }

    settings.endGroup();
}

void ConnectionDialog::loadSettings()
{
    QSettings settings;
    settings.beginGroup("ConnectionDialog");
    resize(settings.value("size", size()).toSize());
    settings.endGroup();
}

void ConnectionDialog::saveSettings()
{
    QSettings settings;
    settings.beginGroup("ConnectionDialog");
    settings.setValue("size", size());
    settings.endGroup();
}

QString ConnectionDialog::buildConnectionName() const
{
    QString host = m_hostEdit->text().trimmed();
    QString db = m_databaseEdit->text().trimmed();
    QString user = m_usernameEdit->text().trimmed();

    if (host.isEmpty() && db.isEmpty()) {
        return QString();
    }

    return QString("%1@%2/%3").arg(user, host, db);
}

void ConnectionDialog::updateUIForDatabaseType(DatabaseType type)
{
    // Set default port for each database type
    switch (type) {
    case DatabaseType::PostgreSQL:
        m_portSpinBox->setValue(5432);
        break;
    case DatabaseType::MySQL:
        m_portSpinBox->setValue(3306);
        break;
    case DatabaseType::SQLite:
    case DatabaseType::DuckDB:
        m_portSpinBox->setValue(0);
        m_hostEdit->setEnabled(false);
        m_portSpinBox->setEnabled(false);
        m_usernameEdit->setEnabled(false);
        m_passwordEdit->setEnabled(false);
        m_sslGroup->setEnabled(false);
        return;
    default:
        m_portSpinBox->setValue(5432);
        break;
    }

    // Enable all fields for server-based databases
    m_hostEdit->setEnabled(true);
    m_portSpinBox->setEnabled(true);
    m_usernameEdit->setEnabled(true);
    m_passwordEdit->setEnabled(true);
    m_sslGroup->setEnabled(true);
}

} // namespace tablepro