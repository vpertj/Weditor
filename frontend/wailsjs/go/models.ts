export namespace main {
	
	export class ExportHTMLResult {
	    success: boolean;
	    filePath: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportHTMLResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.filePath = source["filePath"];
	        this.message = source["message"];
	    }
	}
	export class RecentFile {
	    path: string;
	    name: string;
	    openedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new RecentFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.openedAt = source["openedAt"];
	    }
	}

}

