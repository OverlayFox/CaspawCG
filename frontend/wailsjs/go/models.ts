export namespace data {
	
	export class Data {
	    Key: string;
	    Type: string;
	    Value: any;
	
	    static createFrom(source: any = {}) {
	        return new Data(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Type = source["Type"];
	        this.Value = source["Value"];
	    }
	}
	export class Location {
	    Key: string;
	    Type: string;
	
	    static createFrom(source: any = {}) {
	        return new Location(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Type = source["Type"];
	    }
	}

}

export namespace responses {
	
	export class CINF {
	    Filename: string;
	    Type: string;
	    FileSize: number;
	    // Go type: time
	    LastModified: any;
	    FrameCount: number;
	    FrameRate: types.FrameRate;
	
	    static createFrom(source: any = {}) {
	        return new CINF(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Filename = source["Filename"];
	        this.Type = source["Type"];
	        this.FileSize = source["FileSize"];
	        this.LastModified = this.convertValues(source["LastModified"], null);
	        this.FrameCount = source["FrameCount"];
	        this.FrameRate = this.convertValues(source["FrameRate"], types.FrameRate);
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

export namespace types {
	
	export class FrameRate {
	    Num: number;
	    Den: number;
	
	    static createFrom(source: any = {}) {
	        return new FrameRate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Num = source["Num"];
	        this.Den = source["Den"];
	    }
	}
	export class Sizing {
	    posX: number;
	    posY: number;
	    sizeX: number;
	    sizeY: number;
	
	    static createFrom(source: any = {}) {
	        return new Sizing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.posX = source["posX"];
	        this.posY = source["posY"];
	        this.sizeX = source["sizeX"];
	        this.sizeY = source["sizeY"];
	    }
	}

}

export namespace ui {
	
	export class CGDataGroup {
	    Template: string;
	    Layer: number;
	    Channels: number[];
	    Data: Record<string, any>;
	    Sizing: types.Sizing;
	    Delay: number;
	
	    static createFrom(source: any = {}) {
	        return new CGDataGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Template = source["Template"];
	        this.Layer = source["Layer"];
	        this.Channels = source["Channels"];
	        this.Data = source["Data"];
	        this.Sizing = this.convertValues(source["Sizing"], types.Sizing);
	        this.Delay = source["Delay"];
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
	export class FieldConfig {
	    key: string;
	    type: string;
	    id: string;
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new FieldConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.type = source["type"];
	        this.id = source["id"];
	        this.source = source["source"];
	    }
	}
	export class MediaWidgetConfig {
	    id: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	    name?: string;
	    filename: string;
	    layer: number;
	    channel: number;
	    channelExpr?: string;
	    delay?: number;
	    loop: boolean;
	
	    static createFrom(source: any = {}) {
	        return new MediaWidgetConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	        this.name = source["name"];
	        this.filename = source["filename"];
	        this.layer = source["layer"];
	        this.channel = source["channel"];
	        this.channelExpr = source["channelExpr"];
	        this.delay = source["delay"];
	        this.loop = source["loop"];
	    }
	}
	export class WidgetConfig {
	    id: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	    name?: string;
	    template: string;
	    layer: number;
	    channel: number;
	    channelExpr?: string;
	    posX?: number;
	    posY?: number;
	    sizeX?: number;
	    sizeY?: number;
	    delay?: number;
	    fields: FieldConfig[];
	
	    static createFrom(source: any = {}) {
	        return new WidgetConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	        this.name = source["name"];
	        this.template = source["template"];
	        this.layer = source["layer"];
	        this.channel = source["channel"];
	        this.channelExpr = source["channelExpr"];
	        this.posX = source["posX"];
	        this.posY = source["posY"];
	        this.sizeX = source["sizeX"];
	        this.sizeY = source["sizeY"];
	        this.delay = source["delay"];
	        this.fields = this.convertValues(source["fields"], FieldConfig);
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
	export class GroupConfig {
	    id: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	    name: string;
	    widgets: WidgetConfig[];
	    mediaWidgets?: MediaWidgetConfig[];
	
	    static createFrom(source: any = {}) {
	        return new GroupConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.w = source["w"];
	        this.h = source["h"];
	        this.name = source["name"];
	        this.widgets = this.convertValues(source["widgets"], WidgetConfig);
	        this.mediaWidgets = this.convertValues(source["mediaWidgets"], MediaWidgetConfig);
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
	export class LayoutConfig {
	    version: number;
	    widgets: WidgetConfig[];
	    groups?: GroupConfig[];
	    mediaWidgets?: MediaWidgetConfig[];
	
	    static createFrom(source: any = {}) {
	        return new LayoutConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.widgets = this.convertValues(source["widgets"], WidgetConfig);
	        this.groups = this.convertValues(source["groups"], GroupConfig);
	        this.mediaWidgets = this.convertValues(source["mediaWidgets"], MediaWidgetConfig);
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

