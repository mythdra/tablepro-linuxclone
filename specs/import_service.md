# Import Service (C++20 + Qt 6)

## Overview
The Import Service handles bulk data import from files (SQL dumps, CSV, JSON) into connected databases. It uses streaming parsers for memory efficiency and supports cancellation via `QFuture`.

## 1. Import Pipeline
```cpp
// core/ImportService.hpp
class ImportService : public QObject {
    Q_OBJECT

public:
    explicit ImportService(QObject* parent = nullptr);

    // Main import entry point
    QFuture<ImportResult> importFile(
        const QUuid& connectionId,
        const QString& filePath,
        const QString& format,
        const QString& encoding = "UTF-8"
    );

    // Cancel ongoing import
    void cancelImport();

signals:
    void importProgress(const QUuid& importId, const ImportProgress& progress);
    void importError(const QUuid& importId, const QString& message, int lineNumber);
    void importComplete(const QUuid& importId, int statementsProcessed);

private:
    QString decompressGzip(const QString& path);
    QFuture<ImportResult> executeSqlStream(
        const QUuid& connectionId,
        const QString& filePath,
        const QString& encoding
    );
    QFuture<ImportResult> executeCsvStream(
        const QUuid& connectionId,
        const QString& filePath,
        const QString& targetTable
    );

    QAtomicInt m_cancelRequested;
    QMap<QUuid, QFuture<ImportResult>> m_activeImports;
};
```

```cpp
// core/ImportService.cpp
QFuture<ImportResult> ImportService::importFile(
    const QUuid& connectionId,
    const QString& filePath,
    const QString& format,
    const QString& encoding)
{
    m_cancelRequested.store(0);

    // 1. Decompress if .gz
    QString actualPath = filePath;
    if (filePath.endsWith(".gz")) {
        actualPath = decompressGzip(filePath);
    }

    // 2. Route to appropriate handler
    if (format == "sql") {
        return executeSqlStream(connectionId, actualPath, encoding);
    } else if (format == "csv") {
        return executeCsvStream(connectionId, actualPath, "UTF-8");
    } else if (format == "json") {
        // JSON import handling
    }

    return QtConcurrent::run([]() { return ImportResult{}; });
}

QString ImportService::decompressGzip(const QString& path) {
    // Use zlib directly via Qt
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly)) {
        return {};
    }

    // Skip gzip header and decompress
    QByteArray compressed = file.readAll();
    QByteArray decompressed;

    z_stream stream = {};
    inflateInit2(&stream, 16 + MAX_WBITS);  // Auto-detect gzip

    stream.next_in = reinterpret_cast<Bytef*>(compressed.data());
    stream.avail_in = compressed.size();

    char buffer[32768];
    do {
        stream.next_out = reinterpret_cast<Bytef*>(buffer);
        stream.avail_out = sizeof(buffer);
        int ret = inflate(&stream, Z_NO_FLUSH);
        if (ret != Z_OK && ret != Z_STREAM_END) {
            inflateEnd(&stream);
            return {};
        }
        decompressed.append(buffer, sizeof(buffer) - stream.avail_out);
    } while (stream.avail_in > 0);

    inflateEnd(&stream);

    // Write to temp file
    QTemporaryFile* tempFile = new QTemporaryFile();
    tempFile->setAutoRemove(false);
    tempFile->open();
    tempFile->write(decompressed);
    tempFile->close();

    return tempFile->fileName();  // Caller responsible for cleanup
}
```

## 2. Streaming SQL Parser
```cpp
// core/SqlFileParser.hpp
class SqlFileParser : public QObject {
    Q_OBJECT

public:
    struct Statement {
        QString text;
        int lineNumber;
    };

    explicit SqlFileParser(QObject* parent = nullptr);

    // Streaming parse - emits statements via signal
    void parseFile(const QString& path, const QString& encoding);

    // Cancellation
    void cancel();

signals:
    void statementParsed(const Statement& stmt);
    void parsingFinished();
    void parsingError(const QString& message, int lineNumber);

private:
    enum class State {
        Normal,
        SingleLineComment,
        MultiLineComment,
        SingleQuote,
        DoubleQuote,
        BacktickQuote
    };

    State m_state{State::Normal};
    int m_lineNumber{1};
    int m_statementStartLine{1};
    QTextStream m_stream;
    QString m_currentStatement;
    bool m_hasContent{false};
    QAtomicInt m_cancelRequested;
};
```

```cpp
// core/SqlFileParser.cpp
void SqlFileParser::parseFile(const QString& path, const QString& encoding) {
    QFile file(path);
    if (!file.open(QIODevice::ReadOnly | QIODevice::Text)) {
        emit parsingError(tr("Cannot open file: %1").arg(path), 0);
        return;
    }

    m_stream.setDevice(&file);
    m_stream.setCodec(encoding.toUtf8().constData());
    m_lineNumber = 1;
    m_state = State::Normal;

    QChar ch;
    QChar prevCh;

    while (!m_stream.atEnd()) {
        if (m_cancelRequested.load()) {
            emit parsingError(tr("Parsing cancelled"), m_lineNumber);
            return;
        }

        ch = m_stream.read(1)[0];
        if (ch == '\n') {
            m_lineNumber++;
        }

        switch (m_state) {
            case State::Normal:
                if (ch == '-' && m_stream.peek(1) == '-') {
                    m_state = State::SingleLineComment;
                } else if (ch == '/' && m_stream.peek(1) == '*') {
                    m_state = State::MultiLineComment;
                    m_stream.read(1);  // consume *
                } else if (ch == '\'') {
                    m_state = State::SingleQuote;
                } else if (ch == '"') {
                    m_state = State::DoubleQuote;
                } else if (ch == '`') {
                    m_state = State::BacktickQuote;
                } else if (ch == ';' && m_hasContent) {
                    // End of statement
                    emit statementParsed({
                        m_currentStatement.trimmed(),
                        m_statementStartLine
                    });
                    m_currentStatement.clear();
                    m_hasContent = false;
                    m_statementStartLine = m_lineNumber;
                } else {
                    m_hasContent = m_hasContent || !ch.isSpace();
                }
                break;

            case State::SingleLineComment:
                if (ch == '\n') {
                    m_state = State::Normal;
                }
                break;

            case State::MultiLineComment:
                if (ch == '*' && m_stream.peek(1) == '/') {
                    m_stream.read(1);  // consume /
                    m_state = State::Normal;
                }
                break;

            case State::SingleQuote:
            case State::DoubleQuote:
            case State::BacktickQuote:
                // Track escape sequences and closing quotes
                if (ch == '\\' || (prevCh == ch)) {
                    // Escaped quote - stay in current state
                } else if (
                    (m_state == State::SingleQuote && ch == '\'') ||
                    (m_state == State::DoubleQuote && ch == '"') ||
                    (m_state == State::BacktickQuote && ch == '`')
                ) {
                    m_state = State::Normal;
                }
                break;
        }

        if (!ch.isSpace()) {
            m_currentStatement.append(ch);
        }

        prevCh = ch;
    }

    // Emit final statement if any
    if (m_hasContent) {
        emit statementParsed({
            m_currentStatement.trimmed(),
            m_statementStartLine
        });
    }

    emit parsingFinished();
}
```

## 3. Transaction Execution
```cpp
// core/ImportService.cpp
QFuture<ImportResult> ImportService::executeSqlStream(
    const QUuid& connectionId,
    const QString& filePath,
    const QString& encoding)
{
    return QtConcurrent::run([=]() {
        ImportResult result;

        // Get driver
        auto* driver = ConnectionManager::instance()->getDriver(connectionId);
        if (!driver) {
            return ImportResult{.success = false, .error = tr("No active connection")};
        }

        // Disable FK checks for import
        for (const auto& stmt : driver->dialectInfo().disableFkChecks) {
            driver->execute(stmt);
        }

        // Begin transaction
        driver->beginTransaction();

        // Set up parser
        SqlFileParser parser;
        QEventLoop loop;
        QObject::connect(&parser, &SqlFileParser::parsingFinished,
                         &loop, &QEventLoop::quit);
        QObject::connect(&parser, &SqlFileParser::parsingError,
                         &loop, [&result](const QString& msg, int line) {
            result.success = false;
            result.error = msg;
            result.errorLineNumber = line;
            loop.quit();
        });

        int processed = 0;
        QObject::connect(&parser, &SqlFileParser::statementParsed,
                        [&result, driver, &processed, this](const SqlFileParser::Statement& stmt) {
            if (m_cancelRequested.load()) {
                return;
            }

            if (!stmt.text.isEmpty()) {
                if (auto execResult = driver->execute(stmt.text); !execResult.success) {
                    emit importError(QUuid(), execResult.error, stmt.lineNumber);
                    result.success = false;
                    result.error = execResult.error;
                    result.errorLineNumber = stmt.lineNumber;
                    return;
                }
                processed++;

                // Progress every 100 statements
                if (processed % 100 == 0) {
                    emit importProgress(QUuid(), {
                        .processed = processed,
                        .status = tr("Executing statement %1").arg(processed)
                    });
                }
            }
        });

        parser.parseFile(filePath, encoding);
        loop.exec();

        if (!result.success) {
            driver->rollbackTransaction();
            return result;
        }

        // Commit transaction
        driver->commitTransaction();

        // Re-enable FK checks
        for (const auto& stmt : driver->dialectInfo().enableFkChecks) {
            driver->execute(stmt);
        }

        result.success = true;
        result.processedStatements = processed;
        emit importComplete(QUuid(), processed);

        return result;
    });
}
```

## 4. Progress Throttling
```cpp
// Use Qt's built-in throttling via QTimer
class ImportProgressDialog : public QDialog {
    Q_OBJECT

public:
    explicit ImportProgressDialog(QWidget* parent = nullptr);

    void setFuture(const QFuture<ImportResult>& future);

private slots:
    void onProgress(const ImportProgress& progress);

private:
    QProgressBar* m_progressBar;
    QLabel* m_statusLabel;
    QLabel* m_processedLabel;
    QFuture<ImportResult> m_importFuture;
    QFutureWatcher<ImportResult>* m_watcher;

    // Throttle UI updates to ~15fps
    QElapsedTimer m_lastUpdateTimer;
    static constexpr int kUpdateIntervalMs = 66;
};

void ImportProgressDialog::onProgress(const ImportProgress& progress) {
    // Throttle updates
    if (m_lastUpdateTimer.elapsed() < kUpdateIntervalMs) {
        return;
    }
    m_lastUpdateTimer.restart();

    m_progressBar->setValue(progress.percent());
    m_statusLabel->setText(progress.status);
    m_processedLabel->setText(tr("Processed: %1").arg(progress.processed));
}
```

## 5. Cancellation
```cpp
// Import dialog cancels via QFuture
void ImportProgressDialog::reject() {
    if (m_importFuture.isRunning()) {
        // Request cancellation
        m_importFuture.cancel();

        // Wait for completion (with timeout)
        if (!m_importFuture.waitForFinished(5000)) {
            qWarning() << "Import did not cancel within timeout";
        }
    }
    QDialog::reject();
}

// Parser checks cancellation flag
void SqlFileParser::cancel() {
    m_cancelRequested.store(1);
}

// ImportService propagates cancellation
QFuture<ImportResult> ImportService::importFile(...) {
    m_cancelRequested.store(0);

    return QtConcurrent::run([=]() {
        // Check cancellation periodically
        if (m_cancelRequested.load()) {
            driver->rollbackTransaction();
            return ImportResult{
                .success = false,
                .error = tr("Import cancelled")
            };
        }
        // ... rest of import logic
    });
}
```

## 6. CSV Import Handler
```cpp
// core/CsvImportHandler.hpp
class CsvImportHandler : public QObject {
    Q_OBJECT

public:
    QFuture<ImportResult> importCsv(
        const QUuid& connectionId,
        const QString& filePath,
        const QString& targetTable,
        char delimiter = ',',
        bool hasHeader = true
    );

private:
    QStringList parseLine(const QString& line, char delimiter);
    QString escapeValue(const QVariant& value);
    QVariantMap mapColumns(
        const QStringList& values,
        const QStringList& columnNames,
        const QList<QSqlField>& tableSchema
    );
};
```

## 7. Format Handlers (Plugin-style Registration)
```cpp
// core/ImportFormatRegistry.hpp
class ImportFormat : public QObject {
    Q_OBJECT

public:
    virtual QString id() const = 0;
    virtual QString name() const = 0;
    virtual QStringList extensions() const = 0;
    virtual QFuture<ImportResult> import(
        const QUuid& connectionId,
        const QString& filePath,
        const ImportOptions& options
    ) = 0;
};

class ImportFormatRegistry : public QObject {
    Q_OBJECT
    Q_GLOBAL_STATIC(ImportFormatRegistry, instance)

public:
    static ImportFormatRegistry* instance();

    void registerFormat(ImportFormat* format);
    ImportFormat* getFormat(const QString& extension);
    QStringList supportedExtensions();

private:
    QMap<QString, ImportFormat*> m_formats;
};

// Built-in formats registered at startup
class SqlImportFormat : public ImportFormat { /* ... */ };
class CsvImportFormat : public ImportFormat { /* ... */ };
class JsonImportFormat : public ImportFormat { /* ... */ };
```

## Advantages over Go Implementation

| Aspect | Go | C++20 + Qt |
|--------|-----|------------|
| Streaming | Go channels | Qt signals |
| Goroutines | `go func()` | `QtConcurrent::run()` |
| Cancellation | `context.Context` | `QFuture::cancel()` + atomic flag |
| Transaction | `database/sql` | Driver-specific APIs |
| Gzip | `compress/gzip` | zlib via Qt |
| Progress events | `runtime.EventsEmit()` | Qt signals |
| Memory | GC-managed | RAII + smart pointers |
