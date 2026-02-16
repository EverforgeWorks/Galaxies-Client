import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Ship, GameState, Contract, Planet, ShipModule, TravelEvent, TravelResponse } from '../types';
import { 
  GetShipState, GetPlanets, GetAvailableContracts, GetModules, 
  Travel, AcceptJob, DropJob, Refuel, BuyModule 
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

// CONFIGURATION CONSTANTS (Must match universe.yaml)
const FUEL_MASS_PER_UNIT = 3;

export const useGameStore = defineStore('game', () => {
    // --- STATE ---
    const ship = ref<Ship | null>(null);
    const universe = ref<Planet[]>([]);
    
    const availableJobs = ref<Contract[]>([]);
    const availableModules = ref<ShipModule[]>([]);
    const chatMessages = ref<any[]>([]);

    const arrivalEvents = ref<TravelEvent[]>([]);

    const uiState = ref<GameState>({
        isDocked: false,
        isLoading: false,
        lastError: null,
        showEvents: false
    });

    // --- GETTERS (Computed) ---
    
    // Calculates total physics mass: Base + Fuel + Cargo/Pax
    const totalMass = computed(() => {
        if (!ship.value) return 0;
        
        let mass = ship.value.base_mass;

        // 1. Add Fuel Mass
        mass += (ship.value.fuel * FUEL_MASS_PER_UNIT);

        // 2. Add Contract Mass (Cargo & Passengers)
        if (ship.value.active_contracts) {
            const contractMass = ship.value.active_contracts.reduce((sum, c) => {
                return sum + (c.quantity * c.mass_per_unit);
            }, 0);
            mass += contractMass;
        }

        return mass;
    });

    // --- HELPER ---
    async function performAction(actionName: string, actionFn: () => Promise<any>): Promise<boolean> {
        uiState.value.isLoading = true;
        uiState.value.lastError = null;
        try {
            await actionFn();
            return true;
        } catch (e: any) {
            console.error(`${actionName} failed:`, e);
            uiState.value.lastError = `${actionName} FAILED: ${e}`;
            return false;
        } finally {
            uiState.value.isLoading = false;
        }
    }

    // --- ACTIONS ---

    async function refreshAll() {
        try {
            const [shipData, planets] = await Promise.all([
                GetShipState(),
                universe.value.length === 0 ? GetPlanets() : Promise.resolve(universe.value)
            ]);
            ship.value = shipData as Ship;
            universe.value = planets as Planet[] || [];

            if (ship.value) {
                availableJobs.value = await GetAvailableContracts() as Contract[] || [];
                availableModules.value = await GetModules() as ShipModule[] || [];
            }
        } catch (e) { console.error("Sync Error", e); }
    }

    function initGameEvents() {
        console.log("HOOKING INTO SHIP SYSTEMS...");
        EventsOn("market_pulse", (updatedPlanets: string[]) => {
            console.log("MARKET UPDATE:", updatedPlanets);
            if (!uiState.value.isLoading) refreshAll();
        });
        refreshAll();
    }

    async function travel(destination: string): Promise<{ success: boolean, duration: number }> {
        let duration = 0;
        uiState.value.isLoading = true;
        uiState.value.lastError = null;

        try {
            const response = await Travel(destination);
            
            if (response.success) {
                ship.value = response.ship as Ship;
                arrivalEvents.value = response.events;
                duration = response.duration_seconds;
                return { success: true, duration };
            } else {
                uiState.value.lastError = response.error || null;
                uiState.value.isLoading = false;
                return { success: false, duration: 0 };
            }
        } catch (e) {
            uiState.value.lastError = "NAVIGATION SYSTEM FAILURE";
            uiState.value.isLoading = false;
            return { success: false, duration: 0 };
        }
    }

    function revealEvents() {
        if (arrivalEvents.value.length > 0) {
            uiState.value.showEvents = true;
        }
        uiState.value.isLoading = false;
    }

    function clearEvents() {
        uiState.value.showEvents = false;
        arrivalEvents.value = [];
    }

    async function acceptContract(id: string) {
        const success = await performAction("Accept Contract", async () => {
             const res = await AcceptJob(id);
             if (!res) throw new Error("Contract unavailable or ship full");
        });
        if (success) await refreshAll();
    }

    async function dropContract(id: string) {
        const success = await performAction("Drop Contract", async () => {
             const res = await DropJob(id);
             if (!res) throw new Error("Contract not found");
        });
        if (success) await refreshAll();
    }

    async function refuelShip() {
        const success = await performAction("Refuel", async () => {
             const res = await Refuel();
             if (!res) throw new Error("Insufficient credits or full tank");
        });
        if (success) await refreshAll();
    }

    async function buyModule(key: string) {
        const success = await performAction("Buy Module", async () => {
             const res = await BuyModule(key);
             if (!res) throw new Error("Purchase failed");
        });
        if (success) await refreshAll();
    }

    return {
        ship, universe, availableJobs, availableModules, chatMessages, uiState, arrivalEvents,
        totalMass, // Exported Getter
        refreshAll, initGameEvents, travel, acceptContract, dropContract, refuelShip, buyModule,
        revealEvents, clearEvents
    };
});