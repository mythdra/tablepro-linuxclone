#pragma once

#include <QTableView>
#include <QHeaderView>
#include <QMenu>
#include <QAction>
#include <memory>

namespace tablepro {

class ResultSetModel;

/**
 * High-performance table view for displaying database query results.
 * Supports editing, sorting, filtering, and column management.
 */
class DataGridWidget : public QTableView
{
    Q_OBJECT

public:
    explicit DataGridWidget(QWidget* parent = nullptr);
    ~DataGridWidget() override;

    // Model management
    void setModel(QAbstractItemModel* model) override;
    ResultSetModel* resultSetModel() const;

    // Column management
    void setColumnVisible(int column, bool visible);
    bool isColumnVisible(int column) const;
    void setColumnWidths(const QVector<int>& widths);
    void autoResizeColumns();

    // Selection
    QVector<int> selectedRows() const;
    QVector<int> selectedColumns() const;
    QString selectedAsText() const;

    // Copy/Paste
    void copy();
    void copyWithHeaders();

    // Export
    void exportToCsv(const QString& filePath);
    void exportToJson(const QString& filePath);

signals:
    void selectionChanged(const QModelIndexList& selected);
    void doubleClicked(const QModelIndex& index);
    void contextMenuRequested(const QPoint& pos);

protected:
    void keyPressEvent(QKeyEvent* event) override;
    void contextMenuEvent(QContextMenuEvent* event) override;

private slots:
    void onHeaderContextMenu(const QPoint& pos);
    void onHeaderSectionResized(int logicalIndex, int oldSize, int newSize);

private:
    void setupActions();
    void setupConnections();
    void applyDefaultStyle();

    // Context menu actions
    std::unique_ptr<QMenu> m_headerMenu;
    QAction* m_actionCopy;
    QAction* m_actionCopyWithHeaders;
    QAction* m_actionExportCsv;
    QAction* m_actionExportJson;

    // Hidden columns tracking
    QSet<int> m_hiddenColumns;
};

} // namespace tablepro