import { create } from 'zustand';
import type { CellChange, InsertedRow, DeletedRow, PendingChanges, ChangeSummary } from '../types/change';
import { v4 as uuidv4 } from 'uuid';

/**
 * Change store interface for tracking data grid mutations.
 */
interface ChangeStore {
  // State - changes per tab/query
  cellChanges: Map<string, CellChange[]>;
  insertedRows: Map<string, InsertedRow[]>;
  deletedRows: Map<string, DeletedRow[]>;
  
  // Actions
  updateCell: (tabId: string, rowId: string, column: string, oldValue: any, newValue: any) => void;
  insertRow: (tabId: string, data: Record<string, any>) => void;
  deleteRow: (tabId: string, rowId: string, data: Record<string, any>) => void;
  discardCellChange: (tabId: string, rowId: string, column: string) => void;
  discardAllChanges: (tabId: string) => void;
  getPendingChanges: (tabId: string) => PendingChanges;
  getChangeSummary: (tabId: string) => ChangeSummary;
  hasChanges: (tabId: string) => boolean;
  getCellChange: (tabId: string, rowId: string, column: string) => CellChange | undefined;
}

/**
 * Change store for tracking data grid mutations.
 */
export const useChangeStore = create<ChangeStore>((set, get) => ({
  cellChanges: new Map(),
  insertedRows: new Map(),
  deletedRows: new Map(),

  updateCell: (tabId: string, rowId: string, column: string, oldValue: any, newValue: any) => {
    const tabChanges = get().cellChanges.get(tabId) || [];
    
    // Check if this cell already has a change
    const existingIndex = tabChanges.findIndex(
      (c) => c.rowId === rowId && c.column === column
    );

    const newChange: CellChange = {
      rowId,
      column,
      oldValue,
      newValue,
      timestamp: new Date().toISOString(),
    };

    if (existingIndex >= 0) {
      // Update existing change
      tabChanges[existingIndex] = newChange;
    } else {
      // Add new change
      tabChanges.push(newChange);
    }

    const newCellChanges = new Map(get().cellChanges);
    newCellChanges.set(tabId, tabChanges);
    set({ cellChanges: newCellChanges });
  },

  insertRow: (tabId: string, data: Record<string, any>) => {
    const tabInserts = get().insertedRows.get(tabId) || [];
    
    const newRow: InsertedRow = {
      rowId: uuidv4(),
      data,
      timestamp: new Date().toISOString(),
    };

    tabInserts.push(newRow);

    const newInsertedRows = new Map(get().insertedRows);
    newInsertedRows.set(tabId, tabInserts);
    set({ insertedRows: newInsertedRows });
  },

  deleteRow: (tabId: string, rowId: string, data: Record<string, any>) => {
    const tabDeletes = get().deletedRows.get(tabId) || [];
    
    const newDeleted: DeletedRow = {
      rowId,
      data,
      timestamp: new Date().toISOString(),
    };

    tabDeletes.push(newDeleted);

    const newDeletedRows = new Map(get().deletedRows);
    newDeletedRows.set(tabId, tabDeletes);
    set({ deletedRows: newDeletedRows });
  },

  discardCellChange: (tabId: string, rowId: string, column: string) => {
    const tabChanges = get().cellChanges.get(tabId) || [];
    const filtered = tabChanges.filter(
      (c) => !(c.rowId === rowId && c.column === column)
    );

    const newCellChanges = new Map(get().cellChanges);
    newCellChanges.set(tabId, filtered);
    set({ cellChanges: newCellChanges });
  },

  discardAllChanges: (tabId: string) => {
    const newCellChanges = new Map(get().cellChanges);
    const newInsertedRows = new Map(get().insertedRows);
    const newDeletedRows = new Map(get().deletedRows);
    
    newCellChanges.set(tabId, []);
    newInsertedRows.set(tabId, []);
    newDeletedRows.set(tabId, []);
    
    set({
      cellChanges: newCellChanges,
      insertedRows: newInsertedRows,
      deletedRows: newDeletedRows,
    });
  },

  getPendingChanges: (tabId: string) => {
    return {
      cellChanges: get().cellChanges.get(tabId) || [],
      insertedRows: get().insertedRows.get(tabId) || [],
      deletedRows: get().deletedRows.get(tabId) || [],
    };
  },

  getChangeSummary: (tabId: string) => {
    const changes = get().cellChanges.get(tabId) || [];
    const inserts = get().insertedRows.get(tabId) || [];
    const deletes = get().deletedRows.get(tabId) || [];

    return {
      updates: changes.length,
      inserts: inserts.length,
      deletes: deletes.length,
    };
  },

  hasChanges: (tabId: string) => {
    const summary = get().getChangeSummary(tabId);
    return summary.updates > 0 || summary.inserts > 0 || summary.deletes > 0;
  },

  getCellChange: (tabId: string, rowId: string, column: string) => {
    const changes = get().cellChanges.get(tabId) || [];
    return changes.find((c) => c.rowId === rowId && c.column === column);
  },
}));
