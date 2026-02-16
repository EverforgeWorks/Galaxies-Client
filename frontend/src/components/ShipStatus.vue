<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()

// Computes the display name of the current planet based on the key
const planetName = computed(() => {
    if (!store.universe || !store.ship) return "UNKNOWN"
    const p = store.universe.find(x => x.key === store.ship?.location_key)
    return p ? p.name : store.ship?.location_key
})

// Fuel Calculations
const fuelPct = computed(() => store.ship ? (store.ship.fuel / store.ship.max_fuel) * 100 : 0)
const fuelCost = computed(() => store.ship ? Math.floor((store.ship.max_fuel - store.ship.fuel) / 100 * 4) : 0)

async function handleRefuel() {
    await store.refuelShip()
}
</script>

<template>
  <div class="status-panel">
    <div class="row">
        <span class="ship-name">{{ store.ship?.name || 'NO_SIGNAL' }}</span>
        <span class="loc">@{{ planetName }}</span>
    </div>

    <div class="stats-grid">
        <div class="stat-cell">
            <span class="label">CREDITS</span>
            <span class="value val-cr">{{ store.ship?.credits?.toLocaleString() }}</span>
        </div>
        <div class="stat-cell">
            <span class="label">MASS</span>
            <span class="value">{{ store.totalMass.toLocaleString() }}kg</span>
        </div>
        <div class="stat-cell">
             <span class="label">FUEL ({{ store.ship?.fuel }} / {{ store.ship?.max_fuel }})</span>
             <button class="btn-refuel" 
                :disabled="fuelPct > 99 || store.uiState.isLoading" 
                @click="handleRefuel"
             >
                {{ fuelPct > 99 ? 'FULL' : `FILL -${fuelCost}cr` }}
             </button>
        </div>
    </div>
    
    <div class="fuel-track">
        <div class="fuel-bar" :style="{ width: fuelPct + '%' }" :class="{ 'crit': fuelPct < 20 }"></div>
    </div>
  </div>
</template>

<style scoped>
.status-panel { font-family: 'Courier New', monospace; font-size: 0.8rem; background: #001100; padding: 10px; }

.row { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 6px; }
.ship-name { font-weight: bold; color: #fff; font-size: 0.9rem; }
.loc { color: #00ff41; }

.stats-grid { display: grid; grid-template-columns: 1fr 1fr 1.2fr; gap: 10px; margin-bottom: 5px; }
.stat-cell { display: flex; flex-direction: column; }
.label { font-size: 0.65rem; color: #006600; margin-bottom: 2px; }
.value { font-size: 0.85rem; color: #00aa00; }
.val-cr { color: #fff; font-weight: bold; }

.btn-refuel {
    background: #004400; border: none; color: #00ff41; 
    font-size: 0.7rem; cursor: pointer; padding: 1px 4px;
}
.btn-refuel:hover:not(:disabled) { background: #00ff41; color: #000; }
.btn-refuel:disabled { opacity: 0.3; cursor: default; }

.fuel-track { height: 4px; background: #002200; margin-top: 4px; }
.fuel-bar { height: 100%; background: #00ff41; transition: width 0.5s; }
.fuel-bar.crit { background: #ff0000; }
</style>