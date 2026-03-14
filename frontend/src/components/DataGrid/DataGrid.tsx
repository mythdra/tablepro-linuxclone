import React, { useMemo, useCallback } from 'react';

import { AgGridReact } from '@ag-grid-community/react';
import type {
  ColDef,
  SortChangedEvent,
  ValueFormatterParams,
} from '@ag-grid-community/core';

import '@ag-grid-community/styles/ag-grid.css';
import '@ag-grid-community/styles/ag-theme-alpine.css';

import { NullCellRenderer } from './NullCellRenderer';
import type { ColumnInfo } from '../../types';

export interface SortModel {
  colId: string;
  sort: 'asc' | 'desc';
}

export interface DataGridProps {
  columns: ColumnInfo[];
  rowData: Record<string, unknown>[];
  onSortChanged?: (sortModel: SortModel[]) => void;
  pageSize?: number;
}

function formatDate(value: unknown): string {
  if (value === null || value === undefined) return '';
  const str = String(value);
  const d = new Date(str);
  if (isNaN(d.getTime())) return str;
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function formatNumber(value: unknown): string {
  if (value === null || value === undefined) return '';
  const num = Number(value);
  if (isNaN(num)) return String(value);
  return num.toLocaleString();
}

function formatBoolean(value: unknown): string {
  if (value === null || value === undefined) return '';
  return value ? 'TRUE' : 'FALSE';
}

function getValueFormatter(
  dataType: string | undefined
): ((params: ValueFormatterParams) => string) | undefined {
  switch (dataType) {
    case 'date':
      return (params: ValueFormatterParams) => formatDate(params.value);
    case 'datetime':
      return (params: ValueFormatterParams) => {
        if (params.value === null || params.value === undefined) return '';
        const str = String(params.value);
        const d = new Date(str);
        if (isNaN(d.getTime())) return str;
        const year = d.getFullYear();
        const month = String(d.getMonth() + 1).padStart(2, '0');
        const day = String(d.getDate()).padStart(2, '0');
        const hours = String(d.getHours()).padStart(2, '0');
        const minutes = String(d.getMinutes()).padStart(2, '0');
        const seconds = String(d.getSeconds()).padStart(2, '0');
        return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
      };
    case 'integer':
    case 'float':
    case 'number':
      return (params: ValueFormatterParams) => formatNumber(params.value);
    case 'boolean':
      return (params: ValueFormatterParams) => formatBoolean(params.value);
    default:
      return undefined;
  }
}

function mapColumnsToColDefs(columns: ColumnInfo[]): ColDef[] {
  return columns.map((col) => {
    const colDef: ColDef = {
      field: col.name,
      headerName: col.name,
      sortable: true,
      resizable: true,
      cellRenderer: NullCellRenderer,
    };

    const formatter = getValueFormatter(col.type);
    if (formatter) {
      colDef.valueFormatter = formatter;
    }

    return colDef;
  });
}

function DataGrid({
  columns,
  rowData,
  onSortChanged,
  pageSize = 100,
}: DataGridProps): React.ReactElement {
  const columnDefs = useMemo(() => mapColumnsToColDefs(columns), [columns]);

  const defaultColDef = useMemo<ColDef>(
    () => ({
      sortable: true,
      resizable: true,
      minWidth: 80,
    }),
    []
  );

  const handleSortChanged = useCallback(
    (event: SortChangedEvent) => {
      if (!onSortChanged) return;

      const sortModel: SortModel[] = [];
      const columnState = event.api.getColumnState();

      for (const col of columnState) {
        if (col.sort) {
          sortModel.push({
            colId: col.colId,
            sort: col.sort as 'asc' | 'desc',
          });
        }
      }

      onSortChanged(sortModel);
    },
    [onSortChanged]
  );

  return (
    <div
      className="ag-theme-alpine"
      style={{ width: '100%', height: '100%' }}
    >
      <AgGridReact
        columnDefs={columnDefs}
        rowData={rowData}
        defaultColDef={defaultColDef}
        rowBuffer={20}
        rowHeight={28}
        headerHeight={32}
        onSortChanged={handleSortChanged}
        suppressCellFocus={true}
        animateRows={false}
        suppressColumnVirtualisation={false}
        domLayout={rowData.length < pageSize ? 'autoHeight' : 'normal'}
      />
    </div>
  );
}

export default DataGrid;
