export namespace main {
	
	export class Bookmark {
	    id: string;
	    folder: string;
	    url: string;
	    title: string;
	    localPath: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Bookmark(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.folder = source["folder"];
	        this.url = source["url"];
	        this.title = source["title"];
	        this.localPath = source["localPath"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class SummaryRef {
	    id: string;
	    folder: string;
	    relPath: string;
	    title: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SummaryRef(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.folder = source["folder"];
	        this.relPath = source["relPath"];
	        this.title = source["title"];
	        this.updatedAt = source["updatedAt"];
	    }
	}

}

