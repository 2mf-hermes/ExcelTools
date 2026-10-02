export namespace main {
	
	export class MergeFilesRequest {
	    paths: string[];
	    headerRows: number;
	    sheetIndex: number;
	    columnAlign: string;
	    addSourceFile: boolean;
	    keepBaseOtherSheets: boolean;
	    keepBlankRows: boolean;
	    outputDir: string;
	
	    static createFrom(source: any = {}) {
	        return new MergeFilesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.paths = source["paths"];
	        this.headerRows = source["headerRows"];
	        this.sheetIndex = source["sheetIndex"];
	        this.columnAlign = source["columnAlign"];
	        this.addSourceFile = source["addSourceFile"];
	        this.keepBaseOtherSheets = source["keepBaseOtherSheets"];
	        this.keepBlankRows = source["keepBlankRows"];
	        this.outputDir = source["outputDir"];
	    }
	}
	export class MergeSheetsRequest {
	    sourcePath: string;
	    sheetNames: string[];
	    headerMode: string;
	    addSourceSheet: boolean;
	    skipEmpty: boolean;
	    outputSheetName: string;
	    sourceColumnLabel: string;
	    outputPath: string;
	
	    static createFrom(source: any = {}) {
	        return new MergeSheetsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.sheetNames = source["sheetNames"];
	        this.headerMode = source["headerMode"];
	        this.addSourceSheet = source["addSourceSheet"];
	        this.skipEmpty = source["skipEmpty"];
	        this.outputSheetName = source["outputSheetName"];
	        this.sourceColumnLabel = source["sourceColumnLabel"];
	        this.outputPath = source["outputPath"];
	    }
	}

}

export namespace model {
	
	export class AppInfo {
	    name: string;
	    version: string;
	    platform: string;
	    localOnly: boolean;
	    updateCheck: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.version = source["version"];
	        this.platform = source["platform"];
	        this.localOnly = source["localOnly"];
	        this.updateCheck = source["updateCheck"];
	    }
	}
	export class FileRef {
	    path: string;
	    name: string;
	    size: number;
	    ext: string;
	    status: string;
	    reason?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.ext = source["ext"];
	        this.status = source["status"];
	        this.reason = source["reason"];
	    }
	}
	export class ItemReport {
	    file: string;
	    status: string;
	    rows: number;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.status = source["status"];
	        this.rows = source["rows"];
	        this.message = source["message"];
	    }
	}
	export class MergeResult {
	    success: boolean;
	    partial: boolean;
	    message: string;
	    outputPath: string;
	    filesUsed: number;
	    rows: number;
	    cols: number;
	    durationMs: number;
	    items: ItemReport[];
	
	    static createFrom(source: any = {}) {
	        return new MergeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.partial = source["partial"];
	        this.message = source["message"];
	        this.outputPath = source["outputPath"];
	        this.filesUsed = source["filesUsed"];
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	        this.durationMs = source["durationMs"];
	        this.items = this.convertValues(source["items"], ItemReport);
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
	export class Settings {
	    language: string;
	    theme: string;
	    autoCheckUpdates: boolean;
	    defaults: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.theme = source["theme"];
	        this.autoCheckUpdates = source["autoCheckUpdates"];
	        this.defaults = source["defaults"];
	    }
	}
	export class SheetRef {
	    name: string;
	    rows: number;
	    cols: number;
	    empty: boolean;
	    hidden: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SheetRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.rows = source["rows"];
	        this.cols = source["cols"];
	        this.empty = source["empty"];
	        this.hidden = source["hidden"];
	    }
	}
	export class SheetListResult {
	    sheets: SheetRef[];
	    note?: string;
	
	    static createFrom(source: any = {}) {
	        return new SheetListResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sheets = this.convertValues(source["sheets"], SheetRef);
	        this.note = source["note"];
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
	
	export class UpdateCheckResult {
	    currentVersion: string;
	    latestVersion: string;
	    status: string;
	    message: string;
	    releaseNotes?: string;
	    releaseUrl?: string;
	    assetName?: string;
	    assetSize?: number;
	    canAutoInstall: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateCheckResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.releaseNotes = source["releaseNotes"];
	        this.releaseUrl = source["releaseUrl"];
	        this.assetName = source["assetName"];
	        this.assetSize = source["assetSize"];
	        this.canAutoInstall = source["canAutoInstall"];
	    }
	}
	export class UpdateInstallResult {
	    status: string;
	    code?: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}

}

