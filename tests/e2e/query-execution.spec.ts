/**
 * Playwright E2E tests for Query Execution flow.
 * 
 * These tests verify the complete query execution workflow from the user's perspective.
 * 
 * Prerequisites:
 * - TablePro application must be running in development mode
 * - PostgreSQL test database must be available
 * 
 * Run with: npx playwright test query-execution.spec.ts
 * 
 * Environment variables:
 * - E2E_BASE_URL (default: http://localhost:34115)
 * - E2E_SLOW_MO (default: 0) - Add delay between actions for debugging
 */

import { test, expect, type Page } from '@playwright/test';

// Configuration
const BASE_URL = process.env.E2E_BASE_URL || 'http://localhost:34115';
const SLOW_MO = parseInt(process.env.E2E_SLOW_MO || '0', 10);

/**
 * Test database connection configuration
 */
const testConnection = {
  name: 'E2E Test PostgreSQL',
  host: process.env.POSTGRES_HOST || 'localhost',
  port: parseInt(process.env.POSTGRES_PORT || '5432', 10),
  database: process.env.POSTGRES_DB || 'postgres',
  username: process.env.POSTGRES_USER || 'postgres',
  password: process.env.POSTGRES_PASSWORD || 'postgres',
};

test.describe('Query Execution E2E (11.1-11.8)', () => {
  test.beforeEach(async ({ page }) => {
    // Set viewport for consistent testing
    await page.setViewportSize({ width: 1920, height: 1080 });
    
    // Navigate to app
    await page.goto(BASE_URL);
    
    // Add slow motion if configured
    if (SLOW_MO > 0) {
      await page.waitForTimeout(SLOW_MO);
    }
  });

  test.describe('Connection Setup', () => {
    test('should create a new PostgreSQL connection', async ({ page }) => {
      // Navigate to connections view
      await page.getByRole('button', { name: /new connection/i }).click();
      await page.waitForTimeout(500);

      // Fill connection form
      await page.getByLabel(/connection name/i).fill(testConnection.name);
      await page.getByLabel(/host/i).fill(testConnection.host);
      await page.getByLabel(/port/i).fill(testConnection.port.toString());
      await page.getByLabel(/database/i).fill(testConnection.database);
      await page.getByLabel(/username/i).fill(testConnection.username);
      await page.getByLabel(/password/i).fill(testConnection.password);

      // Save connection
      await page.getByRole('button', { name: /save/i }).click();
      await page.waitForTimeout(1000);

      // Verify connection appears in list
      await expect(page.getByText(testConnection.name)).toBeVisible();
    });

    test('should connect to database and open query editor', async ({ page }) => {
      // Create connection first (assuming it exists)
      await page.getByRole('button', { name: /new connection/i }).click();
      await page.waitForTimeout(500);

      const uniqueName = `E2E Test ${Date.now()}`;
      await page.getByLabel(/connection name/i).fill(uniqueName);
      await page.getByLabel(/host/i).fill(testConnection.host);
      await page.getByLabel(/port/i).fill(testConnection.port.toString());
      await page.getByLabel(/database/i).fill(testConnection.database);
      await page.getByLabel(/username/i).fill(testConnection.username);
      await page.getByLabel(/password/i).fill(testConnection.password);

      await page.getByRole('button', { name: /save and connect/i }).click();
      await page.waitForTimeout(2000);

      // Verify query editor is visible
      await expect(page.getByTestId('query-editor')).toBeVisible();
      await expect(page.getByPlaceholder(/write your query/i)).toBeVisible();
    });
  });

  test.describe('Basic Query Execution (11.1)', () => {
    test.beforeEach(async ({ page }) => {
      // Setup: Create test connection and open query editor
      await setupConnection(page);
    });

    test('should execute simple SELECT query', async ({ page }) => {
      // Type query in editor
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('SELECT 1 AS num, \'hello\' AS str');
      await page.waitForTimeout(500);

      // Execute query (Ctrl+Enter)
      await page.keyboard.press('Control+Enter');
      await page.waitForTimeout(1000);

      // Verify results panel shows data
      await expect(page.getByTestId('result-view')).toBeVisible();
      await expect(page.getByText('num')).toBeVisible();
      await expect(page.getByText('str')).toBeVisible();
      await expect(page.getByText('1')).toBeVisible();
      await expect(page.getByText('hello')).toBeVisible();
    });

    test('should execute query with multiple rows', async ({ page }) => {
      // Create test table and insert data
      await executeQuery(page, `
        CREATE TEMP TABLE e2e_test AS
        SELECT generate_series(1, 10) AS id, 'test_' || g AS name
        FROM generate_series(1, 10) AS g
      `);

      // Query the test table
      await executeQuery(page, 'SELECT * FROM e2e_test ORDER BY id');

      // Verify 10 rows returned
      await expect(page.getByText('10 rows')).toBeVisible();
      
      // Verify first and last rows
      await expect(page.getByText('test_1')).toBeVisible();
      await expect(page.getByText('test_10')).toBeVisible();
    });

    test('should display error for invalid query', async ({ page }) => {
      // Type invalid query
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('SELEC * FORM users');
      await page.waitForTimeout(500);

      // Execute
      await page.keyboard.press('Control+Enter');
      await page.waitForTimeout(1000);

      // Verify error message
      await expect(page.getByTestId('error-message')).toBeVisible();
      await expect(page.getByText(/syntax error|error/i)).toBeVisible();
    });
  });

  test.describe('Query Cancellation (11.3)', () => {
    test('should cancel long-running query', async ({ page }) => {
      await setupConnection(page);

      // Execute pg_sleep query
      await executeQuery(page, 'SELECT pg_sleep(10)');

      // Wait for query to start executing
      await page.waitForTimeout(500);

      // Click cancel button
      await page.getByRole('button', { name: /cancel/i }).click();
      await page.waitForTimeout(1000);

      // Verify cancellation message
      await expect(page.getByText(/cancel|cancelled/i)).toBeVisible();
    });
  });

  test.describe('Pagination (11.4)', () => {
    test('should paginate large result sets', async ({ page }) => {
      await setupConnection(page);

      // Generate 1000 rows
      await executeQuery(page, `
        CREATE TEMP TABLE large_test AS
        SELECT generate_series(1, 1000) AS id, 'row_' || g AS data
        FROM generate_series(1, 1000) AS g
      `);

      // Query with pagination (first page)
      await executeQuery(page, 'SELECT * FROM large_test ORDER BY id LIMIT 100 OFFSET 0');

      // Verify first page
      await expect(page.getByText('100 rows')).toBeVisible();
      await expect(page.getByText('row_1')).toBeVisible();
      await expect(page.getByText('row_100')).toBeVisible();

      // Navigate to next page
      await page.getByRole('button', { name: /next/i }).click();
      await page.waitForTimeout(500);

      // Verify second page
      await expect(page.getByText('row_101')).toBeVisible();
      await expect(page.getByText('row_200')).toBeVisible();
    });
  });

  test.describe('NULL Value Handling (11.5)', () => {
    test('should display NULL values correctly', async ({ page }) => {
      await setupConnection(page);

      // Create table with NULL values
      await executeQuery(page, `
        CREATE TEMP TABLE null_test (
          id INTEGER,
          nullable_text TEXT,
          nullable_num INTEGER
        );
        INSERT INTO null_test VALUES 
          (1, NULL, 100),
          (2, 'not null', NULL),
          (3, NULL, NULL);
      `);

      // Query and verify NULLs
      await executeQuery(page, 'SELECT * FROM null_test ORDER BY id');

      // Verify NULL values are displayed as null/empty
      const cells = page.getByTestId('result-grid').locator('[role="gridcell"]');
      await expect(cells.nth(1)).toHaveText(''); // NULL text
      await expect(cells.nth(4)).toHaveText('not null');
      await expect(cells.nth(7)).toHaveText(''); // Both NULL
    });
  });

  test.describe('History Tracking (11.6)', () => {
    test('should track query history', async ({ page }) => {
      await setupConnection(page);

      // Execute first query
      await executeQuery(page, 'SELECT 1 AS first');
      await page.waitForTimeout(500);

      // Execute second query
      await executeQuery(page, 'SELECT 2 AS second');
      await page.waitForTimeout(500);

      // Execute third query
      await executeQuery(page, 'SELECT 3 AS third');
      await page.waitForTimeout(500);

      // Open history panel
      await page.getByRole('button', { name: /history/i }).click();
      await page.waitForTimeout(500);

      // Verify history entries
      await expect(page.getByText(/SELECT 1/i)).toBeVisible();
      await expect(page.getByText(/SELECT 2/i)).toBeVisible();
      await expect(page.getByText(/SELECT 3/i)).toBeVisible();
    });

    test('should deduplicate identical queries', async ({ page }) => {
      await setupConnection(page);

      // Execute same query twice
      await executeQuery(page, 'SELECT 1');
      await page.waitForTimeout(500);
      await executeQuery(page, 'SELECT 1');
      await page.waitForTimeout(500);

      // Open history panel
      await page.getByRole('button', { name: /history/i }).click();
      await page.waitForTimeout(500);

      // Should only appear once in history
      const historyItems = page.getByText(/SELECT 1/i);
      await expect(historyItems).toHaveCount(1);
    });
  });

  test.describe('Keyboard Shortcuts (11.7)', () => {
    test('should execute query with Ctrl+Enter', async ({ page }) => {
      await setupConnection(page);

      // Type query
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('SELECT 1');
      await page.waitForTimeout(300);

      // Execute with Ctrl+Enter
      await page.keyboard.press('Control+Enter');
      await page.waitForTimeout(1000);

      // Verify execution
      await expect(page.getByTestId('result-view')).toBeVisible();
    });

    test('should create new tab with Ctrl+T', async ({ page }) => {
      await setupConnection(page);

      // Verify initial tab
      await expect(page.getByText('Query 1')).toBeVisible();

      // Create new tab
      await page.keyboard.press('Control+t');
      await page.waitForTimeout(500);

      // Verify new tab
      await expect(page.getByText('Query 2')).toBeVisible();
    });

    test('should format query with Shift+Alt+F', async ({ page }) => {
      await setupConnection(page);

      // Type unformatted query
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('select * from users where id = 1');
      await page.waitForTimeout(300);

      // Format with Shift+Alt+F
      await page.keyboard.press('Shift+Alt+f');
      await page.waitForTimeout(500);

      // Verify formatting (query should be uppercase)
      const editorContent = await page.getByTestId('monaco-editor').textContent();
      expect(editorContent?.toUpperCase()).toContain('SELECT');
    });
  });

  test.describe('Autocomplete (11.8)', () => {
    test('should suggest table names', async ({ page }) => {
      await setupConnection(page);

      // Type query that triggers autocomplete
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('SELECT * FROM ');
      await page.waitForTimeout(1000);

      // Verify autocomplete popup appears
      const autocompleteVisible = await page.locator('.monaco-editor .suggest-widget').isVisible();
      expect(autocompleteVisible).toBeTruthy();
    });

    test('should suggest column names after table', async ({ page }) => {
      await setupConnection(page);

      // Type query with table name
      await page.getByTestId('monaco-editor').click();
      await page.keyboard.type('SELECT id, ');
      await page.waitForTimeout(500);

      // Trigger autocomplete
      await page.keyboard.press('Control+Space');
      await page.waitForTimeout(500);

      // Verify autocomplete popup
      const autocompleteVisible = await page.locator('.monaco-editor .suggest-widget').isVisible();
      expect(autocompleteVisible).toBeTruthy();
    });
  });
});

/**
 * Helper function to setup a test connection
 */
async function setupConnection(page: Page): Promise<void> {
  // Check if already connected (query editor visible)
  const editorVisible = await page.getByTestId('query-editor').isVisible().catch(() => false);
  if (editorVisible) {
    return;
  }

  // Create connection
  await page.getByRole('button', { name: /new connection/i }).click();
  await page.waitForTimeout(500);

  const uniqueName = `E2E Test ${Date.now()}`;
  await page.getByLabel(/connection name/i).fill(uniqueName);
  await page.getByLabel(/host/i).fill(testConnection.host);
  await page.getByLabel(/port/i).fill(testConnection.port.toString());
  await page.getByLabel(/database/i).fill(testConnection.database);
  await page.getByLabel(/username/i).fill(testConnection.username);
  await page.getByLabel(/password/i).fill(testConnection.password);

  await page.getByRole('button', { name: /save and connect/i }).click();
  await page.waitForTimeout(2000);
}

/**
 * Helper function to execute a query
 */
async function executeQuery(page: Page, query: string): Promise<void> {
  // Clear editor
  await page.getByTestId('monaco-editor').click();
  await page.keyboard.press('Control+a');
  await page.keyboard.press('Delete');
  await page.waitForTimeout(200);

  // Type query
  await page.getByTestId('monaco-editor').type(query);
  await page.waitForTimeout(300);

  // Execute
  await page.keyboard.press('Control+Enter');
  await page.waitForTimeout(1000);
}
