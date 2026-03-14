import { useState, useMemo } from 'react';
import Editor from '@monaco-editor/react';
import { Check, X, RotateCcw, AlertCircle, Loader2 } from 'lucide-react';
import { useChangeStore } from '../stores/changeStore';
import type { ColumnMetadata } from '../types/change';

/**
 * Props for the ChangePanel component.
 */
interface ChangePanelProps {
  /** Unique tab identifier */
  tabId: string;
  /** Column metadata for SQL generation */
  columns: ColumnMetadata[];
  /** Table name for SQL generation */
  tableName: string;
  /** Callback to commit changes */
  onCommit: () => Promise<void>;
  /** Callback to discard changes */
  onDiscard: () => void;
  /** Database type for SQL dialect */
  databaseType: string;
}

/**
 * Generates SQL INSERT statements from new rows.
 */
function generateInsertSQL(
  insertedRows: { rowId: string; data: Record<string, any> }[],
  columns: ColumnMetadata[],
  tableName: string,
  dialect: string
): string[] {
  const statements: string[] = [];
  
  // Find auto-increment columns to exclude
  const autoIncrementCols = new Set<string>();
  columns.forEach((col) => {
    // Simple heuristic - exclude primary keys with default values
    if (col.isPrimaryKey && col.defaultValue) {
      autoIncrementCols.add(col.name);
    }
  });

  insertedRows.forEach((row) => {
    const data = row.data;
    const cols: string[] = [];
    const values: string[] = [];

    columns.forEach((col) => {
      // Skip auto-increment columns
      if (autoIncrementCols.has(col.name)) return;
      
      const value = data[col.name];
      // Skip if value is null/undefined and column has default
      if ((value === null || value === undefined) && col.defaultValue) return;

      cols.push(quoteIdentifier(col.name, dialect));
      values.push(formatValue(value, dialect));
    });

    if (cols.length > 0) {
      statements.push(
        `INSERT INTO ${quoteIdentifier(tableName, dialect)}\n(${cols.join(', ')})\nVALUES (${values.join(', ')});`
      );
    }
  });

  return statements;
}

/**
 * Generates SQL DELETE statements from deleted rows.
 */
function generateDeleteSQL(
  deletedRows: { rowId: string; data: Record<string, any> }[],
  pkColumns: ColumnMetadata[],
  tableName: string,
  dialect: string
): string[] {
  const statements: string[] = [];

  deletedRows.forEach((row) => {
    const whereParts: string[] = [];
    pkColumns.forEach((pk) => {
      const pkValue = formatValue(row.data[pk.name], dialect);
      whereParts.push(`${quoteIdentifier(pk.name, dialect)} = ${pkValue}`);
    });

    if (whereParts.length > 0) {
      statements.push(
        `DELETE FROM ${quoteIdentifier(tableName, dialect)}\nWHERE ${whereParts.join(' AND ')};`
      );
    }
  });

  return statements;
}

/**
 * Formats a value for SQL (adds quotes, handles NULL, etc.).
 */
function formatValue(value: any, _dialect: string): string {
  if (value === null || value === undefined || value === '') {
    return 'NULL';
  }

  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value);
  }

  // String value - escape single quotes
  const escaped = String(value).replace(/'/g, "''");
  return `'${escaped}'`;
}

/**
 * Quotes an identifier based on dialect.
 */
function quoteIdentifier(identifier: string, dialect: string): string {
  switch (dialect) {
    case 'mysql':
    case 'duckdb':
      return `\`${identifier}\``;
    case 'mssql':
      return `[${identifier}]`;
    case 'postgres':
    case 'sqlite':
    case 'clickhouse':
    default:
      return `"${identifier}"`;
  }
}

/**
 * ChangePanel component - shows SQL preview with Commit/Discard buttons.
 * 
 * Features:
 * - Monaco Editor for SQL preview with syntax highlighting
 * - Change summary (N UPDATEs, M INSERTs, K DELETEs)
 * - Commit button with loading state
 * - Discard button to cancel all changes
 */
export function ChangePanel({
  tabId,
  columns,
  tableName,
  onCommit,
  onDiscard,
  databaseType,
}: ChangePanelProps) {
  const [isCommitting, setIsCommitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const getPendingChanges = useChangeStore((state) => state.getPendingChanges);
  const getChangeSummary = useChangeStore((state) => state.getChangeSummary);
  const discardAllChanges = useChangeStore((state) => state.discardAllChanges);

  const changes = getPendingChanges(tabId);
  const summary = getChangeSummary(tabId);
  const hasChanges = summary.updates > 0 || summary.inserts > 0 || summary.deletes > 0;

  // Generate SQL preview
  const sqlPreview = useMemo(() => {
    const statements: string[] = [];

    // DELETE statements first (foreign key constraints)
    const pkColumns = columns.filter((c) => c.isPrimaryKey);
    const deleteStatements = generateDeleteSQL(
      changes.deletedRows,
      pkColumns,
      tableName,
      databaseType
    );
    statements.push(...deleteStatements);

    // UPDATE statements second
    // Note: In a real app, we'd need actual row data here
    // For now, generating placeholder
    const updateStatements = changes.cellChanges.map((change) => {
      const pkColumns = columns.filter((c) => c.isPrimaryKey);
      const value = formatValue(change.newValue, databaseType);
      const whereClauses = pkColumns
        .map((pk) => `${quoteIdentifier(pk.name, databaseType)} = ?`)
        .join(' AND ');
      
      return `UPDATE ${quoteIdentifier(tableName, databaseType)}\nSET ${quoteIdentifier(change.column, databaseType)} = ${value}\nWHERE ${whereClauses};`;
    });
    statements.push(...updateStatements);

    // INSERT statements last
    const insertStatements = generateInsertSQL(
      changes.insertedRows,
      columns,
      tableName,
      databaseType
    );
    statements.push(...insertStatements);

    return statements.join('\n\n');
  }, [changes, columns, tableName, databaseType]);

  const handleCommit = async () => {
    setIsCommitting(true);
    setError(null);

    try {
      await onCommit();
      // Changes are discarded after successful commit
      discardAllChanges(tabId);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to commit changes');
    } finally {
      setIsCommitting(false);
    }
  };

  const handleDiscard = () => {
    discardAllChanges(tabId);
    onDiscard();
  };

  // If no changes, show empty state
  if (!hasChanges) {
    return (
      <div className="flex flex-col h-full border-t border-slate-700 bg-slate-800">
        <div className="flex items-center justify-between px-4 py-2 border-b border-slate-700">
          <h3 className="text-sm font-medium text-slate-300">Changes</h3>
        </div>
        <div className="flex-1 flex items-center justify-center text-slate-500">
          <div className="text-center">
            <p className="text-sm">No pending changes</p>
            <p className="text-xs mt-1">Edit cells to see SQL preview</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full border-t border-slate-700 bg-slate-800">
      {/* Header with summary */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-medium text-slate-300">Pending Changes</h3>
          <div className="flex items-center gap-2 text-xs">
            {summary.updates > 0 && (
              <span className="px-2 py-0.5 bg-yellow-500/20 text-yellow-400 rounded-full">
                {summary.updates} UPDATE{summary.updates !== 1 ? 's' : ''}
              </span>
            )}
            {summary.inserts > 0 && (
              <span className="px-2 py-0.5 bg-green-500/20 text-green-400 rounded-full">
                {summary.inserts} INSERT{summary.inserts !== 1 ? 's' : ''}
              </span>
            )}
            {summary.deletes > 0 && (
              <span className="px-2 py-0.5 bg-red-500/20 text-red-400 rounded-full">
                {summary.deletes} DELETE{summary.deletes !== 1 ? 's' : ''}
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleDiscard}
            disabled={isCommitting}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 bg-slate-700 rounded-lg hover:bg-slate-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Discard all changes"
          >
            <RotateCcw className="w-3.5 h-3.5" />
            Discard
          </button>
          <button
            onClick={handleCommit}
            disabled={isCommitting}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            title="Commit all changes"
          >
            {isCommitting ? (
              <>
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
                Committing...
              </>
            ) : (
              <>
                <Check className="w-3.5 h-3.5" />
                Commit
              </>
            )}
          </button>
        </div>
      </div>

      {/* Error display */}
      {error && (
        <div className="mx-4 mt-3 p-2 bg-red-500/10 border border-red-500/30 rounded-lg flex items-start gap-2">
          <AlertCircle className="w-4 h-4 text-red-400 mt-0.5" />
          <div className="flex-1">
            <p className="text-xs text-red-400">{error}</p>
          </div>
          <button
            onClick={() => setError(null)}
            className="text-red-400 hover:text-red-300"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        </div>
      )}

      {/* SQL Preview */}
      <div className="flex-1 min-h-0 flex flex-col">
        <div className="px-4 py-2 text-xs text-slate-400 border-b border-slate-700/50">
          SQL Preview
        </div>
        <div className="flex-1 min-h-0">
          <Editor
            height="100%"
            language="sql"
            value={sqlPreview || '-- No changes to preview'}
            theme="vs-dark"
            options={{
              fontSize: 12,
              fontFamily: "'Fira Code', 'Cascadia Code', Consolas, monospace",
              minimap: { enabled: false },
              scrollBeyondLastLine: false,
              automaticLayout: true,
              tabSize: 2,
              wordWrap: 'on',
              lineNumbers: 'off',
              readOnly: true,
              renderLineHighlight: 'none',
              cursorBlinking: 'solid',
              padding: { top: 8, bottom: 8 },
            }}
          />
        </div>
      </div>
    </div>
  );
}

export default ChangePanel;
