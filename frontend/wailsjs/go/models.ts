export namespace config {
	
	export class Profile {
	    wavelog_url: string;
	    wavelog_key: string;
	    wavelog_id: string;
	    wavelog_radioname: string;
	    wavelog_pmode: boolean;
	    flrig_host: string;
	    flrig_port: string;
	    flrig_ena: boolean;
	    hamlib_host: string;
	    hamlib_port: string;
	    hamlib_ena: boolean;
	    ignore_pwr: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.wavelog_url = source["wavelog_url"];
	        this.wavelog_key = source["wavelog_key"];
	        this.wavelog_id = source["wavelog_id"];
	        this.wavelog_radioname = source["wavelog_radioname"];
	        this.wavelog_pmode = source["wavelog_pmode"];
	        this.flrig_host = source["flrig_host"];
	        this.flrig_port = source["flrig_port"];
	        this.flrig_ena = source["flrig_ena"];
	        this.hamlib_host = source["hamlib_host"];
	        this.hamlib_port = source["hamlib_port"];
	        this.hamlib_ena = source["hamlib_ena"];
	        this.ignore_pwr = source["ignore_pwr"];
	    }
	}
	export class Config {
	    version: number;
	    profile: number;
	    profileNames: string[];
	    udp_enabled: boolean;
	    udp_port: number;
	    profiles: Profile[];
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.profile = source["profile"];
	        this.profileNames = source["profileNames"];
	        this.udp_enabled = source["udp_enabled"];
	        this.udp_port = source["udp_port"];
	        this.profiles = this.convertValues(source["profiles"], Profile);
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

export namespace main {
	
	export class TestResult {
	    success: boolean;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new TestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.reason = source["reason"];
	    }
	}
	export class UDPStatus {
	    enabled: boolean;
	    port: number;
	    running: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UDPStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.port = source["port"];
	        this.running = source["running"];
	    }
	}

}

export namespace wavelog {
	
	export class Station {
	    station_profile_name: string;
	    station_callsign: string;
	    station_id: string;
	
	    static createFrom(source: any = {}) {
	        return new Station(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.station_profile_name = source["station_profile_name"];
	        this.station_callsign = source["station_callsign"];
	        this.station_id = source["station_id"];
	    }
	}

}

