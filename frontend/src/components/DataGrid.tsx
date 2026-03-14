import { useRef, useCallback, useMemo, useEffect } from 'react';
import { AgGridReact } from '@ag-grid-community/react';
import type {
  ColDef,
  GridReadyEvent,
  CellEditingStartedEvent,
  CellEditingStoppedEvent,
  ICellEditorParams,
  RowClassParams,
  ITooltipParams,
} from '@ag-grid-community/core';
import '@ag-grid-community/styles/ag-grid.css';
import '@ag-grid-community/styles/ag-theme-quartz.css';
import { useChangeStore } from '../stores/changeStore';
import type { ColumnMetadata, GridRowData, CellDataType } from '../types/change';

/**
 * Props for the DataGrid component.
 */
interface DataGridProps {
  /** Unique tab identifier for change tracking */
  tabId: string;
  columns: ColumnMetadata[];
  rowData: GridRowData[];
  onCellValueChanged?: (rowId: string, column: string, oldValue: any, newValue: any) => void;
}

/**
 * Validates a value against the expected data type.
 * Returns { valid: true } or { valid: false, error: string }
 */
function validateCellValue(value: any, dataType: CellDataType, nullable: boolean): { valid: boolean; error?: string } {
  // Handle NULL values
  if (value === null || value === undefined || value === '') {
    if (nullable) {
      return { valid: true };
    }
    return { valid: false, error: 'NULL values not allowed' };
  }

  switch (dataType) {
    case 'integer': {
      const num = Number(value);
      if (Number.isNaN(num) || !Number.isInteger(num)) {
        return { valid: false, error: 'Invalid integer value' };
      }
      return { valid: true };
    }

    case 'float': {
      const num = Number(value);
      if (Number.isNaN(num)) {
        return { valid: false, error: 'Invalid numeric value' };
      }
      return { valid: true };
    }

    case 'boolean': {
      if (typeof value === 'boolean') {
        return { valid: true };
      }
      const str = String(value).toLowerCase();
      if (['true', 'false', '1', '0', 'yes', 'no'].includes(str)) {
        return { valid: true };
      }
      return { valid: false, error: 'Invalid boolean value' };
    }

    case 'date': {
      const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
      if (!dateRegex.test(String(value))) {
        return { valid: false, error: 'Invalid date format. Use YYYY-MM-DD' };
      }
      const date = new Date(String(value));
      if (isNaN(date.getTime())) {
        return { valid: false, error: 'Invalid date' };
      }
      return { valid: true };
    }

    case 'timestamp': {
      const timestampRegex = /^\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}/;
      if (!timestampRegex.test(String(value))) {
        return { valid: false, error: 'Invalid timestamp format. Use YYYY-MM-DD HH:MM:SS' };
      }
      return { valid: true };
    }

    case 'text':
    case 'json':
    case 'blob':
    default:
      return { valid: true };
  }
}

/**
 * Custom text editor cell component for inline editing.
 */
function TextCellEditor(props: ICellEditorParams) {
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, []);

  return (
    <input
      ref={inputRef}
      type="text"
      defaultValue={props.value ?? ''}
      className="w-full h-full px-2 py-1 border-0 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      style={{ fontSize: '13px' }}
    />
  );
}

/**
 * Custom number editor cell component for integer/float editing.
 */
function NumberCellEditor(props: ICellEditorParams) {
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, []);

  return (
    <input
      ref={inputRef}
      type="number"
      defaultValue={props.value ?? ''}
      className="w-full h-full px-2 py-1 border-0 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      style={{ fontSize: '13px' }}
    />
  );
}

/**
 * Custom date editor cell component for date editing.
 */
function DateCellEditor(props: ICellEditorParams) {
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  return (
    <input
      ref={inputRef}
      type="date"
      defaultValue={props.value ?? ''}
      className="w-full h-full px-2 py-1 border-0 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      style={{ fontSize: '13px' }}
    />
  );
}

/**
 * Custom boolean dropdown editor cell component.
 */
function BooleanCellEditor(props: ICellEditorParams) {
  const selectRef = useRef<HTMLSelectElement>(null);

  useEffect(() => {
    if (selectRef.current) {
      selectRef.current.focus();
    }
  }, []);

  return (
    <select
      ref={selectRef}
      defaultValue={props.value ? 'true' : 'false'}
      className="w-full h-full px-2 py-1 border-0 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
      style={{ fontSize: '13px' }}
    >
      <option value="true">TRUE</option>
      <option value="false">FALSE</option>
    </select>
  );
}

/**
 * DataGrid component with inline editing support.
 * 
 * Features:
 * - Double-click or Enter key to activate edit mode
 * - Data type validation for edited values
 * - Primary key column edit prevention
 * - NULL value editing
 * - Escape key to cancel edit
 * - Visual change indicators (yellow for modified cells)
 * - Change tracking via Zustand store
 */
export function DataGrid({ tabId, columns, rowData, onCellValueChanged }: DataGridProps) {
  const gridRef = useRef<AgGridReact>(null);
  const updateCell = useChangeStore((state) => state.updateCell);
  const getCellChange = useChangeStore((state) => state.getCellChange);

  /**
   * Get editor component based on data type.
   */
  const getCellEditor = useCallback((dataType: CellDataType) => {
    switch (dataType) {
      case 'integer':
      case 'float':
        return NumberCellEditor;
      case 'date':
      case 'timestamp':
        return DateCellEditor;
      case 'boolean':
        return BooleanCellEditor;
      default:
        return TextCellEditor;
    }
  }, []);

  /**
   * Column definitions with editing support.
   */
  const columnDefs: ColDef[] = useMemo(() => {
    return columns.map((col): ColDef => {
      const isPrimaryKey = col.isPrimaryKey;

      return {
        field: col.name,
        headerName: col.name,
        headerTooltip: isPrimaryKey ? 'Primary Key (read-only)' : col.type,
        editable: (params) => {
          if (isPrimaryKey) {
            return false;
          }
          if (params.data._isNew) {
            return true;
          }
          if (params.data._isDeleted) {
            return false;
          }
          return true;
        },
        cellEditor: getCellEditor(col.dataType ?? 'text'),
        cellEditorPopup: false,
        valueGetter: (params) => {
          const value = params.data[col.name];
          // Display NULL values as empty string for editing
          if (value === null || value === undefined) {
            return '';
          }
          return value;
        },
        cellClassRules: {
          // Primary key styling
          'bg-slate-100 cursor-not-allowed': () => !!isPrimaryKey,
          // Modified cell styling
          'bg-yellow-100': (params) => {
            const change = getCellChange(tabId, params.data._rowId, col.name);
            return !!change;
          },
          // Deleted row styling
          'bg-red-100 line-through': (params) => params.data._isDeleted,
          // New row styling
          'bg-green-50': (params) => params.data._isNew,
        },
        tooltipValueGetter: (params: ITooltipParams<any, any>) => {
          if (isPrimaryKey) {
            return 'Primary key columns cannot be edited';
          }
          if (params.data._isDeleted) {
            return 'Deleted rows cannot be edited';
          }
          const change = getCellChange(tabId, params.data._rowId, col.name);
          if (change) {
            return `Changed from: ${change.oldValue ?? 'NULL'}\nNew value: ${change.newValue ?? 'NULL'}`;
          }
          return `${col.type}${col.nullable ? ' (nullable)' : ''}`;
        },
        cellStyle: { cursor: isPrimaryKey ? 'not-allowed' : 'default' },
        sortable: true,
        filter: true,
        resizable: true,
        minWidth: 100,
      };
    });
  }, [columns, tabId, getCellChange, getCellEditor]);

  /**
   * Handle cell editing started.
   */
  const onCellEditingStarted = useCallback((_event: CellEditingStartedEvent) => {
    // Focus is handled by the editor component
  }, []);

  /**
   * Handle cell editing stopped.
   */
  const onCellEditingStopped = useCallback((event: CellEditingStoppedEvent) => {
    if (!event.data || !event.colDef.field) {
      return;
    }

    const rowId = event.data._rowId;
    const column = event.colDef.field as string;
    const oldValue = event.oldValue;
    const newValue = event.newValue;

    // Skip if no actual change
    if (oldValue === newValue) {
      return;
    }

    // Find column metadata
    const columnMeta = columns.find((c) => c.name === column);
    if (!columnMeta) {
      return;
    }

    // Validate the new value
    const validation = validateCellValue(newValue, columnMeta.dataType ?? 'text', columnMeta.nullable);
    if (!validation.valid) {
      // Show error toast (simplified - in real app use toast library)
      alert(validation.error);
      // Revert to old value by calling the callback with old value
      if (onCellValueChanged) {
        onCellValueChanged(rowId, column, oldValue, oldValue);
      }
      return;
    }

    // Track the change
    updateCell(tabId, rowId, column, oldValue, newValue);

    // Notify parent component
    if (onCellValueChanged) {
      onCellValueChanged(rowId, column, oldValue, newValue);
    }
  }, [columns, tabId, updateCell, onCellValueChanged]);

  /**
   * Row class rules for visual styling.
   */
  const getRowClass = useCallback((params: RowClassParams) => {
    if (params.data._isDeleted) {
      return 'bg-red-100';
    }
    if (params.data._isNew) {
      return 'bg-green-50';
    }
    return '';
  }, []);

  /**
   * Handle grid ready event.
   */
  const onGridReady = useCallback((_params: GridReadyEvent) => {
    // Grid is ready
  }, []);

  /**
   * Handle double click on cell to activate editing.
   */
  const onCellDoubleClicked = useCallback((params: any) => {
    if (params.column?.colDef.editable && typeof params.column.colDef.editable === 'function') {
      const isEditable = params.column.colDef.editable(params);
      if (isEditable) {
        gridRef.current?.api.startEditingCell({
          rowIndex: params.rowIndex,
          colKey: params.column.colId,
        });
      }
    }
  }, []);

  /**
   * Handle key press events for Enter to edit.
   */
  const onCellKeyDown = useCallback((params: any) => {
    if (params.event.key === 'Enter' && !params.node.editing) {
      if (params.column?.colDef.editable) {
        const isEditable = typeof params.column.colDef.editable === 'function'
          ? params.column.colDef.editable(params)
          : params.column.colDef.editable;
        
        if (isEditable) {
          gridRef.current?.api.startEditingCell({
            rowIndex: params.node.rowIndex,
            colKey: params.column.colId,
          });
          params.event.preventDefault();
        }
      }
    }
    // Escape to cancel editing is handled by AG Grid by default
  }, []);

  // Apply auto-sizing on first render
  useEffect(() => {
    if (gridRef.current?.api) {
      gridRef.current.api.autoSizeAllColumns();
    }
  }, []);

  return (
    <div className="h-full w-full ag-theme-quartz" style={{ '--ag-font-size': '13px' } as any}>
      <AgGridReact
        ref={gridRef}
        rowData={rowData}
        columnDefs={columnDefs}
        defaultColDef={{
          editable: true,
          sortable: true,
          filter: true,
          resizable: true,
          minWidth: 80,
        }}
        rowSelection="single"
        getRowClass={getRowClass}
        onGridReady={onGridReady}
        onCellEditingStarted={onCellEditingStarted}
        onCellEditingStopped={onCellEditingStopped}
        onCellDoubleClicked={onCellDoubleClicked}
        onCellKeyDown={onCellKeyDown}
        stopEditingWhenCellsLoseFocus={true}
        ensureDomOrder={true}
        enableCellTextSelection={true}
        pagination={true}
        paginationPageSize={100}
        domLayout="normal"
      />
    </div>
  );
}

export default DataGrid;
