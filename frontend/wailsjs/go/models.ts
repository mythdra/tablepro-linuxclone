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

