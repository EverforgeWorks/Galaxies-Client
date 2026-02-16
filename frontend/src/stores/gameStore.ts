import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Ship, GameState, Contract, Planet, ShipModule, TravelEvent, TravelResponse } from '../types';
// These imports will work once you run 'wails dev' and the bindings are generated
import { 
  GetShipState, GetPlanets, GetAvailableContracts, GetModules, 
  Travel, AcceptJob, DropJob, Refuel, BuyModule 
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

export const useGameStore = defineStore('game', () => {
    // --- STATE ---
    const ship = ref<Ship | null>(null);
    const universe = ref<Planet[]>([]);
    
    const availableJobs = ref<Contract[]>([]);
    const availableModules = ref<ShipModule[]>([]);
    const chatMessages = ref<any[]>([]);

    // New: Store the latest events to be displayed in the popup
    const arrivalEvents = ref<TravelEvent[]>([]);

    const uiState = ref<GameState>({
        isDocked: false,
        isLoading: false,
        lastError: null,
        showEvents: false
    });

    // --- HELPER: Returns TRUE if successful, FALSE if failed ---
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

    // NEW: Listen for Wails Events (Replaces WebSockets)
    function initGameEvents() {
        console.log("HOOKING INTO SHIP SYSTEMS...");
        
        // Listen for the Go 'market_pulse' event from app.go
        EventsOn("market_pulse", (updatedPlanets: string[]) => {
            console.log("MARKET UPDATE:", updatedPlanets);
            // Refresh data if we are idle
            if (!uiState.value.isLoading) refreshAll();
        });

        // Initial Load
        refreshAll();
    }

    // UPDATED: Calls the Wails App.Travel method directly
    async function travel(destination: string): Promise<{ success: boolean, duration: number }> {
        let duration = 0;
        
        uiState.value.isLoading = true;
        uiState.value.lastError = null;

        try {
            const response = await Travel(destination);
            
            if (response.success) {
                // FORCE CAST: Treat the response as compatible with our Ship type
                ship.value = response.ship as Ship;
                arrivalEvents.value = response.events;
                duration = response.duration_seconds;
                
                return { success: true, duration };
            } else {
                // FIX: Convert 'undefined' to 'null' for strict TS compliance
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

    // Triggered by UI when ready to show the events (after animation)
    function revealEvents() {
        if (arrivalEvents.value.length > 0) {
            uiState.value.showEvents = true;
        }
        uiState.value.isLoading = false; // Release lock
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
        refreshAll, initGameEvents, travel, acceptContract, dropContract, refuelShip, buyModule,
        revealEvents, clearEvents
    };
});