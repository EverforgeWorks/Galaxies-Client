import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { Ship, GameState, Contract, Planet, ShipModule } from '../types';
import { 
  GetShipState, GetAvailableContracts, Travel, AcceptJob, 
  DropJob, GetPlanets, Refuel, GetModules, BuyModule 
} from '../../wailsjs/go/main/App';

export const useGameStore = defineStore('game', () => {
    // --- STATE ---
    const ship = ref<Ship | null>(null);
    const universe = ref<Planet[]>([]);
    
    const availableJobs = ref<Contract[]>([]);
    const availableModules = ref<ShipModule[]>([]);
    const chatMessages = ref<any[]>([]);

    const uiState = ref<GameState>({
        isDocked: false,
        isLoading: false,
        lastError: null
    });

    let socket: WebSocket | null = null;
    let reconnectAttempts = 0;

    // --- HELPER: Returns TRUE if successful, FALSE if failed ---
    async function performAction(actionName: string, actionFn: () => Promise<any>): Promise<boolean> {
        uiState.value.isLoading = true;
        uiState.value.lastError = null;
        try {
            await actionFn();
            // Refresh logic usually happens inside the specific actions or here
            return true;
        } catch (e: any) {
            console.error(`${actionName} failed:`, e);
            // Parse common HTTP codes if they appear in the error string
            if (e.toString().includes("402")) uiState.value.lastError = "INSUFFICIENT FUNDS/FUEL";
            else if (e.toString().includes("409")) uiState.value.lastError = "CONFLICT: CARGO FULL OR JOB GONE";
            else uiState.value.lastError = `${actionName} FAILED: ${e}`;
            
            return false;
        } finally {
            uiState.value.isLoading = false;
        }
    }

    // --- ACTIONS ---

    async function refreshAll() {
        // (Keep your existing refresh logic, but handle errors silently for background refresh)
        try {
            const [shipData, planets] = await Promise.all([
                GetShipState(),
                universe.value.length === 0 ? GetPlanets() : Promise.resolve(universe.value)
            ]);
            ship.value = shipData;
            universe.value = planets || [];

            if (ship.value) {
                availableJobs.value = await GetAvailableContracts() || [];
                availableModules.value = await GetModules() || [];
            }
        } catch (e) { console.error("Sync Error", e); }
    }

    // Updated: Returns Promise<boolean>
    async function travel(destination: string): Promise<boolean> {
        return performAction("Navigation", async () => {
            ship.value = await Travel(destination);
            // We do NOT refresh market immediately here to allow animation to play out
            // The component will trigger the market refresh after arrival
        });
    }

    async function acceptContract(id: string) {
        const success = await performAction("Accept Contract", () => AcceptJob(id));
        if (success) await refreshAll();
    }

    async function dropContract(id: string) {
        const success = await performAction("Drop Contract", () => DropJob(id));
        if (success) await refreshAll();
    }

    async function refuelShip() {
        const success = await performAction("Refuel", () => Refuel());
        if (success) await refreshAll();
    }

    async function buyModule(key: string) {
        const success = await performAction("Buy Module", () => BuyModule(key));
        if (success) await refreshAll();
    }

    // --- WEBSOCKETS (Keep existing implementation) ---
    function connectSocket() { /* ... keep existing ... */ }
    function sendChatMessage(text: string) { /* ... keep existing ... */ }

    return {
        ship, universe, availableJobs, availableModules, chatMessages, uiState,
        refreshAll, travel, acceptContract, dropContract, refuelShip, buyModule,
        connectSocket, sendChatMessage
    };
});