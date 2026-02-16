import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { Ship, GameState, Contract, Planet, ShipModule, TravelEvent } from '../types';
import { 
  GetShipState, GetPlanets, GetAvailableContracts, GetModules, 
  Travel, AcceptJob, DropJob, Refuel, BuyModule, LoadGame, CreateNewGame
} from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const FUEL_MASS_PER_UNIT = 3;

export const useGameStore = defineStore('game', () => {
    // --- STATE ---
    const ship = ref<Ship | null>(null);
    const universe = ref<Planet[]>([]);
    const availableJobs = ref<Contract[]>([]);
    const availableModules = ref<ShipModule[]>([]);
    const arrivalEvents = ref<TravelEvent[]>([]);

    const activeSlot = ref<number | null>(null);
    const currentView = ref<'menu' | 'onboarding' | 'game'>('menu');

    const uiState = ref<GameState>({
        isDocked: false,
        isLoading: false,
        lastError: null,
        showEvents: false
    });

    // --- GETTERS ---
    const totalMass = computed(() => {
        if (!ship.value) return 0;
        let mass = ship.value.base_mass;
        mass += (ship.value.fuel * FUEL_MASS_PER_UNIT);
        if (ship.value.active_contracts) {
            mass += ship.value.active_contracts.reduce((sum, c) => sum + (c.quantity * c.mass_per_unit), 0);
        }
        return mass;
    });

    // --- ACTIONS ---

    async function startNewSession(slot: number, playerName: string, shipName: string, shipType: string) {
        uiState.value.isLoading = true;
        try {
            const res = await CreateNewGame({
                slot,
                player_name: playerName,
                ship_name: shipName,
                ship_type_key: shipType
            });
            if (res.startsWith("CREATION FAILED")) throw new Error(res);
            
            activeSlot.value = slot;
            await refreshAll();
            currentView.value = 'game';
        } catch (e: any) {
            uiState.value.lastError = e.message;
        } finally {
            uiState.value.isLoading = false;
        }
    }

    async function loadSession(slot: number) {
        uiState.value.isLoading = true;
        try {
            const res = await LoadGame(slot);
            if (res.startsWith("LOAD FAILED")) throw new Error(res);
            
            activeSlot.value = slot;
            await refreshAll();
            currentView.value = 'game';
        } catch (e: any) {
            uiState.value.lastError = e.message;
        } finally {
            uiState.value.isLoading = false;
        }
    }

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
        EventsOn("market_pulse", (updatedPlanets: string[]) => {
            if (!uiState.value.isLoading && currentView.value === 'game') refreshAll();
        });
    }

    async function travel(destination: string): Promise<{ success: boolean, duration: number }> {
        uiState.value.isLoading = true;
        try {
            const response = await Travel(destination);
            if (response.success) {
                ship.value = response.ship as Ship;
                arrivalEvents.value = response.events;
                return { success: true, duration: response.duration_seconds };
            } else {
                uiState.value.lastError = response.error || null;
                return { success: false, duration: 0 };
            }
        } catch (e) {
            uiState.value.lastError = "NAVIGATION SYSTEM FAILURE";
            return { success: false, duration: 0 };
        } finally {
            uiState.value.isLoading = false;
        }
    }

    return {
        ship, universe, availableJobs, availableModules, uiState, arrivalEvents,
        activeSlot, currentView, totalMass,
        refreshAll, initGameEvents, travel, startNewSession, loadSession,
        revealEvents: () => { uiState.value.showEvents = true; },
        clearEvents: () => { uiState.value.showEvents = false; arrivalEvents.value = []; }
    };
});