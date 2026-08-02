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
	
	export class FieldSubscription {
	    Source: string;
	    Type: string;
	    InputType: string;
	    Location: string;
	    Range: string;
	    Offset: number;
	
	    static createFrom(source: any = {}) {
	        return new FieldSubscription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Source = source["Source"];
	        this.Type = source["Type"];
	        this.InputType = source["InputType"];
	        this.Location = source["Location"];
	        this.Range = source["Range"];
	        this.Offset = source["Offset"];
	    }
	}
	export class FieldSubscriptionResult {
	    Identifier: string;
	    Value: any;
	    Error: string;
	
	    static createFrom(source: any = {}) {
	        return new FieldSubscriptionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Identifier = source["Identifier"];
	        this.Value = source["Value"];
	        this.Error = source["Error"];
	    }
	}
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
	export class LiteralField {
	    CasparKey: string;
	    Type: string;
	    RawValue: string;
	
	    static createFrom(source: any = {}) {
	        return new LiteralField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CasparKey = source["CasparKey"];
	        this.Type = source["Type"];
	        this.RawValue = source["RawValue"];
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
	    ChannelExpr: string;
	    Fields: types.LiteralField[];
	    Sizing: types.Sizing;
	    DelayMs: number;
	
	    static createFrom(source: any = {}) {
	        return new CGDataGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Template = source["Template"];
	        this.Layer = source["Layer"];
	        this.ChannelExpr = source["ChannelExpr"];
	        this.Fields = this.convertValues(source["Fields"], types.LiteralField);
	        this.Sizing = this.convertValues(source["Sizing"], types.Sizing);
	        this.DelayMs = source["DelayMs"];
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
	    inputType: string;
	    location?: string;
	    source?: string;
	    value?: string;
	    range?: string;
	    offset?: number;
	
	    static createFrom(source: any = {}) {
	        return new FieldConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.type = source["type"];
	        this.inputType = source["inputType"];
	        this.location = source["location"];
	        this.source = source["source"];
	        this.value = source["value"];
	        this.range = source["range"];
	        this.offset = source["offset"];
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
	    channelExpr?: string;
	    delayMs?: number;
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
	        this.channelExpr = source["channelExpr"];
	        this.delayMs = source["delayMs"];
	        this.loop = source["loop"];
	    }
	}
	export class TemplateConfig {
	    id: string;
	    x: number;
	    y: number;
	    w: number;
	    h: number;
	    name?: string;
	    template: string;
	    channelExpr?: string;
	    layer: number;
	    sizing: types.Sizing;
	    delayMs?: number;
	    updateIntervalMs?: number;
	    fields: FieldConfig[];
	
	    static createFrom(source: any = {}) {
	        return new TemplateConfig(source);
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
	        this.channelExpr = source["channelExpr"];
	        this.layer = source["layer"];
	        this.sizing = this.convertValues(source["sizing"], types.Sizing);
	        this.delayMs = source["delayMs"];
	        this.updateIntervalMs = source["updateIntervalMs"];
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
	    widgets: TemplateConfig[];
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
	        this.widgets = this.convertValues(source["widgets"], TemplateConfig);
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
	    widgets: TemplateConfig[];
	    groups?: GroupConfig[];
	    mediaWidgets?: MediaWidgetConfig[];
	
	    static createFrom(source: any = {}) {
	        return new LayoutConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.widgets = this.convertValues(source["widgets"], TemplateConfig);
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
	export class MediaGroupItem {
	    Filename: string;
	    Layer: number;
	    ChannelExpr: string;
	    Loop: boolean;
	    DelayMs: number;
	
	    static createFrom(source: any = {}) {
	        return new MediaGroupItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Filename = source["Filename"];
	        this.Layer = source["Layer"];
	        this.ChannelExpr = source["ChannelExpr"];
	        this.Loop = source["Loop"];
	        this.DelayMs = source["DelayMs"];
	    }
	}
	
	export class RangeField {
	    CasparKey: string;
	    Type: string;
	    Source: string;
	    Range: string;
	    Offset: number;
	
	    static createFrom(source: any = {}) {
	        return new RangeField(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.CasparKey = source["CasparKey"];
	        this.Type = source["Type"];
	        this.Source = source["Source"];
	        this.Range = source["Range"];
	        this.Offset = source["Offset"];
	    }
	}

}

