export namespace change {
	
	export class CellChange {
	    rowIndex: number;
	    column: string;
	    originalValue: any;
	    newValue: any;
	    primaryKey: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new CellChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rowIndex = source["rowIndex"];
	        this.column = source["column"];
	        this.originalValue = source["originalValue"];
	        this.newValue = source["newValue"];
	        this.primaryKey = source["primaryKey"];
	    }
	}
	export class ChangeSummary {
	    updates: number;
	    inserts: number;
	    deletes: number;
	
	    static createFrom(source: any = {}) {
	        return new ChangeSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.updates = source["updates"];
	        this.inserts = source["inserts"];
	        this.deletes = source["deletes"];
	    }
	}
	export class DeletedRow {
	    primaryKey: Record<string, any>;
	    rowIndex: number;
	
	    static createFrom(source: any = {}) {
	        return new DeletedRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.primaryKey = source["primaryKey"];
	        this.rowIndex = source["rowIndex"];
	    }
	}
	export class InsertedRow {
	    tempId: string;
	    data: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new InsertedRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tempId = source["tempId"];
	        this.data = source["data"];
	    }
	}
	export class PendingChanges {
	    cellChanges: CellChange[];
	    insertedRows: InsertedRow[];
	    deletedRows: DeletedRow[];
	    summary: ChangeSummary;
	
	    static createFrom(source: any = {}) {
	        return new PendingChanges(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cellChanges = this.convertValues(source["cellChanges"], CellChange);
	        this.insertedRows = this.convertValues(source["insertedRows"], InsertedRow);
	        this.deletedRows = this.convertValues(source["deletedRows"], DeletedRow);
	        this.summary = this.convertValues(source["summary"], ChangeSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace connection {
	
	export class SSLConfig {
	    enabled: boolean;
	    mode: string;
	    caCert: string;
	    clientCert: string;
	    serverName: string;
	
	    static createFrom(source: any = {}) {
	        return new SSLConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.mode = source["mode"];
	        this.caCert = source["caCert"];
	        this.clientCert = source["clientCert"];
	        this.serverName = source["serverName"];
	    }
	}
	export class SSHTunnelConfig {
	    enabled: boolean;
	    host: string;
	    port: number;
	    username: string;
	    authMethod: string;
	
	    static createFrom(source: any = {}) {
	        return new SSHTunnelConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.authMethod = source["authMethod"];
	    }
	}
	export class DatabaseConnection {
	    id: string;
	    name: string;
	    type: string;
	    group: string;
	    colorTag: string;
	    host: string;
	    port: number;
	    database: string;
	    username: string;
	    localFilePath: string;
	    ssh: SSHTunnelConfig;
	    ssl: SSLConfig;
	    safeMode: string;
	    startupCommand: string;
	    preConnectScript: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.group = source["group"];
	        this.colorTag = source["colorTag"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.database = source["database"];
	        this.username = source["username"];
	        this.localFilePath = source["localFilePath"];
	        this.ssh = this.convertValues(source["ssh"], SSHTunnelConfig);
	        this.ssl = this.convertValues(source["ssl"], SSLConfig);
	        this.safeMode = source["safeMode"];
	        this.startupCommand = source["startupCommand"];
	        this.preConnectScript = source["preConnectScript"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	export class TestConnectionResult {
	    success: boolean;
	    message: string;
	    responseTimeMs?: number;
	
	    static createFrom(source: any = {}) {
	        return new TestConnectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.responseTimeMs = source["responseTimeMs"];
	    }
	}

}

export namespace query {
	
	export class ColumnInfo {
	    name: string;
	    type: string;
	    dataType: string;
	    nullable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.dataType = source["dataType"];
	        this.nullable = source["nullable"];
	    }
	}
	export class CountResult {
	    count: number;
	    isExact: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CountResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.isExact = source["isExact"];
	    }
	}
	export class ResultSet {
	    columns: ColumnInfo[];
	    rows: any[][];
	    rowCount: number;
	    queryTime: number;
	    statement: string;
	    hasMore: boolean;
	    multipleResultSets?: ResultSet[];
	
	    static createFrom(source: any = {}) {
	        return new ResultSet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.columns = this.convertValues(source["columns"], ColumnInfo);
	        this.rows = source["rows"];
	        this.rowCount = source["rowCount"];
	        this.queryTime = source["queryTime"];
	        this.statement = source["statement"];
	        this.hasMore = source["hasMore"];
	        this.multipleResultSets = this.convertValues(source["multipleResultSets"], ResultSet);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StatementResult {
	    statement: string;
	    resultSet?: ResultSet;
	    rowsAffected?: number;
	    error?: string;
	    success: boolean;
	    duration: number;
	
	    static createFrom(source: any = {}) {
	        return new StatementResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statement = source["statement"];
	        this.resultSet = this.convertValues(source["resultSet"], ResultSet);
	        this.rowsAffected = source["rowsAffected"];
	        this.error = source["error"];
	        this.success = source["success"];
	        this.duration = source["duration"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MultiStatementResult {
	    queryId: number[];
	    totalDuration: number;
	    results: StatementResult[];
	    partialFail: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MultiStatementResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.queryId = source["queryId"];
	        this.totalDuration = source["totalDuration"];
	        this.results = this.convertValues(source["results"], StatementResult);
	        this.partialFail = source["partialFail"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaginationContext {
	    page: number;
	    pageSize: number;
	    totalCount: number;
	    totalPages: number;
	    hasNext: boolean;
	    hasPrev: boolean;
	    offset: number;
	    isExact: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PaginationContext(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	        this.totalCount = source["totalCount"];
	        this.totalPages = source["totalPages"];
	        this.hasNext = source["hasNext"];
	        this.hasPrev = source["hasPrev"];
	        this.offset = source["offset"];
	        this.isExact = source["isExact"];
	    }
	}
	export class QueryHistoryEntry {
	    ID: number[];
	    Query: string;
	    // Go type: time
	    Timestamp: any;
	    Duration: number;
	    Success: boolean;
	    Error: string;
	    RowCount: number;
	    Connection: number[];
	
	    static createFrom(source: any = {}) {
	        return new QueryHistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Query = source["Query"];
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.Duration = source["Duration"];
	        this.Success = source["Success"];
	        this.Error = source["Error"];
	        this.RowCount = source["RowCount"];
	        this.Connection = source["Connection"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class QueryResult {
	    resultSet?: ResultSet;
	    queryId: number[];
	    duration: number;
	    pagination?: PaginationContext;
	
	    static createFrom(source: any = {}) {
	        return new QueryResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.resultSet = this.convertValues(source["resultSet"], ResultSet);
	        this.queryId = source["queryId"];
	        this.duration = source["duration"];
	        this.pagination = this.convertValues(source["pagination"], PaginationContext);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

