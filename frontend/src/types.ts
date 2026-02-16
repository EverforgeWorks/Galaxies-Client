// frontend/src/types.ts

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
    type: "cargo" | "passenger";
    item_name: string;
    quantity: number;
    payout: number;
    origin_key: string;
    destination_key: string;
    mass_per_unit: number;
}

export interface Ship {
    name: string;
    credits: number;
    fuel: number;
    max_fuel: number;
    location_key: string;
    
    // New Engine Stats
    base_burn_rate: number;
    burn_damping: number;
    base_mass: number;

    cargo_capacity: number;
    passenger_slots: number;
    
    installed_modules: ShipModule[];
    active_contracts: Contract[];
}

export interface Planet {
    key: string;
    name: string;
    coordinates: number[];
}

// Added to satisfy the store's need for "GameState"
export interface GameState {
    isDocked: boolean;
    isLoading: boolean;
    lastError: string | null;
}

export type Player = Ship;