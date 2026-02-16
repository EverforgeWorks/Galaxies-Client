export namespace game {
	
	export class Contract {
	    id: string;
	    type: string;
	    item_name: string;
	    item_key: string;
	    quantity: number;
	    mass_per_unit: number;
	    origin_key: string;
	    destination_key: string;
	    payout: number;
	
	    static createFrom(source: any = {}) {
	        return new Contract(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.item_name = source["item_name"];
	        this.item_key = source["item_key"];
	        this.quantity = source["quantity"];
	        this.mass_per_unit = source["mass_per_unit"];
	        this.origin_key = source["origin_key"];
	        this.destination_key = source["destination_key"];
	        this.payout = source["payout"];
	    }
	}
	export class Planet {
	    key: string;
	    name: string;
	    coordinates: number[];
	    production: string[];
	    demand: string[];
	    min_cargo: number;
	    max_cargo: number;
	    min_passengers: number;
	    max_passengers: number;
	
	    static createFrom(source: any = {}) {
	        return new Planet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.coordinates = source["coordinates"];
	        this.production = source["production"];
	        this.demand = source["demand"];
	        this.min_cargo = source["min_cargo"];
	        this.max_cargo = source["max_cargo"];
	        this.min_passengers = source["min_passengers"];
	        this.max_passengers = source["max_passengers"];
	    }
	}
	export class ShipModule {
	    key: string;
	    name: string;
	    description: string;
	    cost: number;
	    stat_modifier: string;
	    stat_value: number;
	
	    static createFrom(source: any = {}) {
	        return new ShipModule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.cost = source["cost"];
	        this.stat_modifier = source["stat_modifier"];
	        this.stat_value = source["stat_value"];
	    }
	}
	export class Ship {
	    name: string;
	    location_key: string;
	    credits: number;
	    fuel: number;
	    max_fuel: number;
	    base_burn_rate: number;
	    burn_damping: number;
	    base_mass: number;
	    cargo_capacity: number;
	    passenger_slots: number;
	    max_module_slots: number;
	    installed_modules: ShipModule[];
	    active_contracts: Contract[];
	    total_mass: number;
	    current_burn: number;
	
	    static createFrom(source: any = {}) {
	        return new Ship(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.location_key = source["location_key"];
	        this.credits = source["credits"];
	        this.fuel = source["fuel"];
	        this.max_fuel = source["max_fuel"];
	        this.base_burn_rate = source["base_burn_rate"];
	        this.burn_damping = source["burn_damping"];
	        this.base_mass = source["base_mass"];
	        this.cargo_capacity = source["cargo_capacity"];
	        this.passenger_slots = source["passenger_slots"];
	        this.max_module_slots = source["max_module_slots"];
	        this.installed_modules = this.convertValues(source["installed_modules"], ShipModule);
	        this.active_contracts = this.convertValues(source["active_contracts"], Contract);
	        this.total_mass = source["total_mass"];
	        this.current_burn = source["current_burn"];
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
	
	export class TravelEvent {
	    type: string;
	    description: string;
	    effect: string;
	
	    static createFrom(source: any = {}) {
	        return new TravelEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.description = source["description"];
	        this.effect = source["effect"];
	    }
	}

}

export namespace main {
	
	export class TravelQuoteResponse {
	    distance: number;
	    fuel_cost: number;
	    can_afford: boolean;
	    burn_rate: number;
	    estimated_duration_seconds: number;
	
	    static createFrom(source: any = {}) {
	        return new TravelQuoteResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.distance = source["distance"];
	        this.fuel_cost = source["fuel_cost"];
	        this.can_afford = source["can_afford"];
	        this.burn_rate = source["burn_rate"];
	        this.estimated_duration_seconds = source["estimated_duration_seconds"];
	    }
	}
	export class TravelResponse {
	    success: boolean;
	    ship: game.Ship;
	    events: game.TravelEvent[];
	    duration_seconds: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TravelResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.ship = this.convertValues(source["ship"], game.Ship);
	        this.events = this.convertValues(source["events"], game.TravelEvent);
	        this.duration_seconds = source["duration_seconds"];
	        this.error = source["error"];
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

