export namespace database {
	
	export class CharsetInfo {
	    name: string;
	    description: string;
	    collations: string[];
	
	    static createFrom(source: any = {}) {
	        return new CharsetInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.collations = source["collations"];
	    }
	}
	export class ColumnInfo {
	    Name: string;
	    Type: string;
	    Length: number;
	    Nullable: boolean;
	    IsPrimary: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ColumnInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Type = source["Type"];
	        this.Length = source["Length"];
	        this.Nullable = source["Nullable"];
	        this.IsPrimary = source["IsPrimary"];
	    }
	}
	export class CreateDatabaseOptions {
	    name: string;
	    charset: string;
	    collation: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateDatabaseOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.charset = source["charset"];
	        this.collation = source["collation"];
	    }
	}
	export class DatabaseConfig {
	    Type: string;
	    Host: string;
	    Port: number;
	    User: string;
	    Password: string;
	    Database: string;
	    SSLMode: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Host = source["Host"];
	        this.Port = source["Port"];
	        this.User = source["User"];
	        this.Password = source["Password"];
	        this.Database = source["Database"];
	        this.SSLMode = source["SSLMode"];
	    }
	}
	export class DatabaseInfo {
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new DatabaseInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	    }
	}
	export class SchemaInfo {
	    Name: string;
	
	    static createFrom(source: any = {}) {
	        return new SchemaInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	    }
	}
	export class TableInfo {
	    Name: string;
	    Comment: string;
	
	    static createFrom(source: any = {}) {
	        return new TableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Comment = source["Comment"];
	    }
	}

}

