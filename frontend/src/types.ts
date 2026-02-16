export interface ShipModule {
    key: string;
    name: string;
    description: string;
    cost: number;
    stat_modifier: string;
    stat_value: number;
}

export interface Contract {
    id: string;
    // CHANGED: Relaxed from "cargo" | "passenger" to string to match Go backend
    type: string; 
    item_name: string;
    quantity: number;
    payout: number;
    origin_key: string;
    destination_key: string;
    mass_per_unit: number;
}

export interface TravelEvent {
    type: string;
    description: string;
    effect: string;
}

export interface TravelResponse {
    success: boolean;
    ship: Ship;
    events: TravelEvent[];
    duration_seconds: number;
    // ADDED: Optional operator (?) to handle omitempty
    error?: string; 
}

export interface Ship {
    name: string;
    credits: number;
    fuel: number;
    max_fuel: number;
    location_key: string;
    
    base_burn_rate: number;
    burn_damping: number;
    base_mass: number;

    cargo_capacity: number;
    passenger_slots: number;
    max_module_slots: number;
    installed_modules: ShipModule[];
    active_contracts: Contract[];
}

export interface Planet {
    key: string;
    name: string;
    coordinates: number[];
}

export interface GameState {
    isDocked: boolean;
    isLoading: boolean;
    lastError: string | null;
    showEvents: boolean;
}