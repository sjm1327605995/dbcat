// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {database} from '../models';

export function CreateConnection(arg1:database.DatabaseConfig):Promise<string>;

export function CreateDatabase(arg1:database.DatabaseConfig,arg2:database.CreateDatabaseOptions):Promise<void>;

export function ExecuteQuery(arg1:database.DatabaseConfig,arg2:string,arg3:string):Promise<Array<{[key: string]: string}>>;

export function GetDatabaseCharsets(arg1:database.DatabaseConfig):Promise<Array<database.CharsetInfo>>;

export function GetDatabases(arg1:database.DatabaseConfig):Promise<Array<database.DatabaseInfo>>;

export function GetSchemas(arg1:database.DatabaseConfig,arg2:string):Promise<Array<database.SchemaInfo>>;

export function GetTableData(arg1:database.DatabaseConfig,arg2:string,arg3:string,arg4:number,arg5:number):Promise<Array<{[key: string]: string}>>;

export function GetTableRowCount(arg1:database.DatabaseConfig,arg2:string,arg3:string):Promise<number>;

export function GetTableStructure(arg1:database.DatabaseConfig,arg2:string,arg3:string):Promise<Array<database.ColumnInfo>>;

export function GetTables(arg1:database.DatabaseConfig,arg2:string,arg3:string):Promise<Array<database.TableInfo>>;

export function TestConnection(arg1:database.DatabaseConfig):Promise<void>;
