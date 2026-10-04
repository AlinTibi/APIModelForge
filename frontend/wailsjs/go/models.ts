export namespace main {
	
	export class GenerateRequest {
	    json: string;
	    rootName: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.json = source["json"];
	        this.rootName = source["rootName"];
	    }
	}
	export class GeneratedFile {
	    language: string;
	    label: string;
	    code: string;
	    fileName: string;
	    modelCount: number;
	
	    static createFrom(source: any = {}) {
	        return new GeneratedFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language = source["language"];
	        this.label = source["label"];
	        this.code = source["code"];
	        this.fileName = source["fileName"];
	        this.modelCount = source["modelCount"];
	    }
	}
	export class GenerateResponse {
	    success: boolean;
	    error?: string;
	    line?: number;
	    column?: number;
	    files?: GeneratedFile[];
	
	    static createFrom(source: any = {}) {
	        return new GenerateResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.error = source["error"];
	        this.line = source["line"];
	        this.column = source["column"];
	        this.files = this.convertValues(source["files"], GeneratedFile);
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
	
	export class OpenFileResult {
	    cancelled: boolean;
	    content?: string;
	    fileName?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new OpenFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.content = source["content"];
	        this.fileName = source["fileName"];
	        this.error = source["error"];
	    }
	}
	export class SaveFileResult {
	    cancelled: boolean;
	    path?: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveFileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cancelled = source["cancelled"];
	        this.path = source["path"];
	        this.error = source["error"];
	    }
	}
	export class ValidationResult {
	    valid: boolean;
	    message?: string;
	    error?: string;
	    line?: number;
	    column?: number;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.message = source["message"];
	        this.error = source["error"];
	        this.line = source["line"];
	        this.column = source["column"];
	    }
	}

}

