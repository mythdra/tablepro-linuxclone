#include "data_grid_widget.h"
#include "result_set_model.h"

#include <QApplication>
#include <QClipboard>
#include <QKeyEvent>
#include <QFileDialog>
#include <QFile>
#include <QTextStream>
#include <QHeaderView>
#include <QScrollBar>

namespace tablepro {

DataGridWidget::DataGridWidget(QWidget* parent)
    : QTableView(parent)
    , m_headerMenu(std::make_unique<QMenu>(this))
{
    setupActions();
    setupConnections();
    applyDefaultStyle();
}

DataGridWidget::~DataGridWidget() = default;

void DataGridWidget::setModel(QAbstractItemModel* model)
{
    QTableView::setModel(model);
    applyDefaultStyle();
}

ResultSetModel* DataGridWidget::resultSetModel() const
{
    return qobject_cast<ResultSetModel*>(model());
}

void DataGridWidget::setColumnVisible(int column, bool visible)
{
    if (visible) {
        showColumn(column);
        m_hiddenColumns.remove(column);
    } else {
        hideColumn(column);
        m_hiddenColumns.insert(column);
    }
}

bool DataGridWidget::isColumnVisible(int column) const
{
    return !m_hiddenColumns.contains(column);
}

void DataGridWidget::setColumnWidths(const QVector<int>& widths)
{
    for (int i = 0; i < widths.size() && i < model()->columnCount(); ++i) {
        setColumnWidth(i, widths[i]);
    }
}

void DataGridWidget::autoResizeColumns()
{
    resizeColumnsToContents();

    // Set minimum widths
    for (int i = 0; i < model()->columnCount(); ++i) {
        if (columnWidth(i) < 80) {
            setColumnWidth(i, 80);
        }
    }
}

QVector<int> DataGridWidget::selectedRows() const
{
    QSet<int> rows;
    const auto indexes = selectedIndexes();
    for (const QModelIndex& index : indexes) {
        rows.insert(index.row());
    }
    return QVector<int>(rows.begin(), rows.end());
}

QVector<int> DataGridWidget::selectedColumns() const
{
    QSet<int> cols;
    const auto indexes = selectedIndexes();
    for (const QModelIndex& index : indexes) {
        cols.insert(index.column());
    }
    return QVector<int>(cols.begin(), cols.end());
}

QString DataGridWidget::selectedAsText() const
{
    QString text;
    QTextStream stream(&text);

    const auto indexes = selectedIndexes();
    if (indexes.isEmpty()) {
        return text;
    }

    // Get bounds
    int minRow = INT_MAX, maxRow = INT_MIN;
    int minCol = INT_MAX, maxCol = INT_MIN;

    for (const QModelIndex& index : indexes) {
        minRow = qMin(minRow, index.row());
        maxRow = qMax(maxRow, index.row());
        minCol = qMin(minCol, index.column());
        maxCol = qMax(maxCol, index.column());
    }

    // Build text
    for (int row = minRow; row <= maxRow; ++row) {
        for (int col = minCol; col <= maxCol; ++col) {
            if (col > minCol) {
                stream << "\t";
            }
            QModelIndex index = model()->index(row, col);
            if (indexes.contains(index)) {
                stream << model()->data(index).toString();
            }
        }
        stream << "\n";
    }

    stream.flush();
    return text;
}

void DataGridWidget::copy()
{
    const QString text = selectedAsText();
    if (!text.isEmpty()) {
        QApplication::clipboard()->setText(text);
    }
}

void DataGridWidget::copyWithHeaders()
{
    QString text;
    QTextStream stream(&text);

    const auto indexes = selectedIndexes();
    if (indexes.isEmpty()) {
        return;
    }

    // Get bounds
    int minRow = INT_MAX, maxRow = INT_MIN;
    int minCol = INT_MAX, maxCol = INT_MIN;

    for (const QModelIndex& index : indexes) {
        minRow = qMin(minRow, index.row());
        maxRow = qMax(maxRow, index.row());
        minCol = qMin(minCol, index.column());
        maxCol = qMax(maxCol, index.column());
    }

    // Headers
    for (int col = minCol; col <= maxCol; ++col) {
        if (col > minCol) {
            stream << "\t";
        }
        stream << model()->headerData(col, Qt::Horizontal).toString();
    }
    stream << "\n";

    // Data
    for (int row = minRow; row <= maxRow; ++row) {
        for (int col = minCol; col <= maxCol; ++col) {
            if (col > minCol) {
                stream << "\t";
            }
            QModelIndex index = model()->index(row, col);
            if (indexes.contains(index)) {
                stream << model()->data(index).toString();
            }
        }
        stream << "\n";
    }

    stream.flush();
    QApplication::clipboard()->setText(text);
}

void DataGridWidget::exportToCsv(const QString& filePath)
{
    auto* rsModel = resultSetModel();
    if (!rsModel) {
        return;
    }

    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
        return;
    }

    QTextStream stream(&file);
    stream << rsModel->toCsv();
    file.close();
}

void DataGridWidget::exportToJson(const QString& filePath)
{
    auto* rsModel = resultSetModel();
    if (!rsModel) {
        return;
    }

    QFile file(filePath);
    if (!file.open(QIODevice::WriteOnly | QIODevice::Text)) {
        return;
    }

    QTextStream stream(&file);
    stream << rsModel->toJson();
    file.close();
}

void DataGridWidget::keyPressEvent(QKeyEvent* event)
{
    // Ctrl+C - Copy
    if (event->modifiers() == Qt::ControlModifier && event->key() == Qt::Key_C) {
        copy();
        return;
    }

    // Ctrl+Shift+C - Copy with headers
    if (event->modifiers() == (Qt::ControlModifier | Qt::ShiftModifier) && event->key() == Qt::Key_C) {
        copyWithHeaders();
        return;
    }

    // Ctrl+A - Select all
    if (event->modifiers() == Qt::ControlModifier && event->key() == Qt::Key_A) {
        selectAll();
        return;
    }

    QTableView::keyPressEvent(event);
}

void DataGridWidget::contextMenuEvent(QContextMenuEvent* event)
{
    const QPoint pos = event->globalPos();

    // Check if click is in header
    const QPoint headerPos = horizontalHeader()->mapFromGlobal(pos);
    if (horizontalHeader()->rect().contains(headerPos)) {
        onHeaderContextMenu(horizontalHeader()->logicalIndexAt(headerPos));
        return;
    }

    emit contextMenuRequested(pos);
}

void DataGridWidget::onHeaderContextMenu(int logicalIndex)
{
    m_headerMenu->clear();

    // Column visibility
    QAction* hideAction = m_headerMenu->addAction(tr("Hide column"));
    connect(hideAction, &QAction::triggered, this, [this, logicalIndex]() {
        setColumnVisible(logicalIndex, false);
    });

    m_headerMenu->addSeparator();

    // Auto-size
    QAction* autoSizeAction = m_headerMenu->addAction(tr("Auto-size"));
    connect(autoSizeAction, &QAction::triggered, this, [this, logicalIndex]() {
        resizeColumnToContents(logicalIndex);
    });

    QAction* autoSizeAllAction = m_headerMenu->addAction(tr("Auto-size all"));
    connect(autoSizeAllAction, &QAction::triggered, this, &DataGridWidget::autoResizeColumns);

    m_headerMenu->addSeparator();

    // Export
    m_headerMenu->addAction(m_actionExportCsv);
    m_headerMenu->addAction(m_actionExportJson);

    m_headerMenu->popup(QCursor::pos());
}

void DataGridWidget::onHeaderSectionResized(int logicalIndex, int oldSize, int newSize)
{
    Q_UNUSED(logicalIndex)
    Q_UNUSED(oldSize)
    Q_UNUSED(newSize)
    // Could emit signal for persisting column widths
}

void DataGridWidget::setupActions()
{
    m_actionCopy = new QAction(tr("Copy"), this);
    m_actionCopy->setShortcut(QKeySequence::Copy);
    connect(m_actionCopy, &QAction::triggered, this, &DataGridWidget::copy);

    m_actionCopyWithHeaders = new QAction(tr("Copy with Headers"), this);
    m_actionCopyWithHeaders->setShortcut(QKeySequence(Qt::CTRL | Qt::SHIFT | Qt::Key_C));
    connect(m_actionCopyWithHeaders, &QAction::triggered, this, &DataGridWidget::copyWithHeaders);

    m_actionExportCsv = new QAction(tr("Export to CSV..."), this);
    connect(m_actionExportCsv, &QAction::triggered, this, [this]() {
        const QString filePath = QFileDialog::getSaveFileName(this, tr("Export to CSV"),
            QString(), "CSV Files (*.csv)");
        if (!filePath.isEmpty()) {
            exportToCsv(filePath);
        }
    });

    m_actionExportJson = new QAction(tr("Export to JSON..."), this);
    connect(m_actionExportJson, &QAction::triggered, this, [this]() {
        const QString filePath = QFileDialog::getSaveFileName(this, tr("Export to JSON"),
            QString(), "JSON Files (*.json)");
        if (!filePath.isEmpty()) {
            exportToJson(filePath);
        }
    });
}

void DataGridWidget::setupConnections()
{
    connect(horizontalHeader(), &QHeaderView::customContextMenuRequested,
            this, &DataGridWidget::onHeaderContextMenu);

    connect(this, &QTableView::doubleClicked, this, &DataGridWidget::doubleClicked);
}

void DataGridWidget::applyDefaultStyle()
{
    // Enable sorting
    setSortingEnabled(true);

    // Selection behavior
    setSelectionBehavior(QAbstractItemView::SelectItems);
    setSelectionMode(QAbstractItemView::ExtendedSelection);

    // Alternating row colors
    setAlternatingRowColors(true);

    // Header settings
    horizontalHeader()->setStretchLastSection(true);
    horizontalHeader()->setHighlightSections(false);
    horizontalHeader()->setSectionsClickable(true);
    horizontalHeader()->setSectionsMovable(true);

    verticalHeader()->setVisible(true);
    verticalHeader()->setDefaultSectionSize(24);

    // Grid style
    setShowGrid(true);
    setGridStyle(Qt::SolidLine);

    // Edit trigger
    setEditTriggers(QAbstractItemView::DoubleClicked | QAbstractItemView::EditKeyPressed);

    // Word wrap
    setWordWrap(false);
}

} // namespace tablepro