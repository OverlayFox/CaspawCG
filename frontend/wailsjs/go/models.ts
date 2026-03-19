export namespace types {
	
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

