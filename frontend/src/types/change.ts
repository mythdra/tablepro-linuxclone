// Change tracking types for data grid editing
import type { ColumnMetadata as BaseColumnMetadata } from '../types';

export interface CellChange {
  rowId: string;
  column: string;
  oldValue: any;
  newValue: any;
  timestamp: string;
}

export interface InsertedRow {
  rowId: string;
  data: Record<string, any>;
  timestamp: string;
}

export interface DeletedRow {
  rowId: string;
  data: Record<string, any>;
  timestamp: string;
}

export interface PendingChanges {
  cellChanges: CellChange[];
  insertedRows: InsertedRow[];
  deletedRows: DeletedRow[];
}

export interface ChangeSummary {
  updates: number;
  inserts: number;
  deletes: number;
}

export type CellDataType = 
  | 'text'
  | 'integer'
  | 'float'
  | 'boolean'
  | 'date'
  | 'timestamp'
  | 'json'
  | 'blob';

/** Extended column metadata with data type info for validation */
export interface ColumnMetadata extends BaseColumnMetadata {
  dataType?: CellDataType;
}

export interface GridRowData {
  _rowId: string;
  _isNew?: boolean;
  _isDeleted?: boolean;
  [key: string]: any;
}
