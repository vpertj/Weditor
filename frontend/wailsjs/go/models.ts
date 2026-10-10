export namespace main {
	
	export class BuildInfo {
	    version: string;
	    commit: string;
	    buildTime: string;
	    platform: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.buildTime = source["buildTime"];
	        this.platform = source["platform"];
	    }
	}
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
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    channel: string;
	    publishedAt: string;
	    notes: string;
	    assetName: string;
	    assetSize: number;
	    assetUrl: string;
	    sha256: string;
	    reason: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.channel = source["channel"];
	        this.publishedAt = source["publishedAt"];
	        this.notes = source["notes"];
	        this.assetName = source["assetName"];
	        this.assetSize = source["assetSize"];
	        this.assetUrl = source["assetUrl"];
	        this.sha256 = source["sha256"];
	        this.reason = source["reason"];
	        this.error = source["error"];
	    }
	}
	export class UpdateResult {
	    success: boolean;
	    message: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.filePath = source["filePath"];
	    }
	}
	export class UpdateSettings {
	    channel: string;
	    autoCheck: boolean;
	    skippedVersion: string;
	    lastCheckAt: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.channel = source["channel"];
	        this.autoCheck = source["autoCheck"];
	        this.skippedVersion = source["skippedVersion"];
	        this.lastCheckAt = source["lastCheckAt"];
	    }
	}

}

