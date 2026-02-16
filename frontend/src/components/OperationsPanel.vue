<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'
import type { Contract } from '../types'

const store = useGameStore()

// --- TAB STATE ---
type Tab = 'cargo' | 'passengers' | 'modules'
const activeTab = ref<Tab>('cargo')

// --- HELPER: Planet Name Lookup ---
function getPlanetName(key: string): string {
    const p = store.universe.find(p => p.key === key)
    return p ? p.name : key
}

// --- HELPER: Sorter ---
// Sorts by Destination Name (A-Z), then Payout (High-to-Low)
function contractSorter(a: Contract, b: Contract) {
    const nameA = getPlanetName(a.destination_key)
    const nameB = getPlanetName(b.destination_key)
    
    if (nameA !== nameB) {
        return nameA.localeCompare(nameB)
    }
    return b.payout - a.payout
}

// --- COMPUTED DATA HELPERS ---

// Cargo Logic
const onboardCargo = computed(() => 
    (store.ship?.active_contracts.filter(c => c.type === 'cargo') || []).sort(contractSorter)
)
const marketCargo = computed(() => 
    (store.availableJobs.filter(c => c.type === 'cargo') || []).sort(contractSorter)
)

// Passenger Logic
const onboardPassengers = computed(() => 
    (store.ship?.active_contracts.filter(c => c.type === 'passenger') || []).sort(contractSorter)
)
const marketPassengers = computed(() => 
    (store.availableJobs.filter(c => c.type === 'passenger') || []).sort(contractSorter)
)

// Module Logic
const installedModules = computed(() => 
    store.ship?.installed_modules || []
)
const marketModules = computed(() => 
    store.availableModules || []
)

// --- ACTIONS ---
function handleAccept(id: string) {
    store.acceptContract(id)
}

function handleDrop(id: string) {
    if(confirm("Jettison this contract? You will not be paid.")) {
        store.dropContract(id)
    }
}

function handleBuyModule(key: string) {
    if(confirm("Purchase and install this module?")) {
        store.buyModule(key)
    }
}
</script>

<template>
  <div class="ops-panel">
    
    <div class="tabs-header">
      <button class="tab-btn" :class="{ active: activeTab === 'cargo' }" @click="activeTab = 'cargo'">CARGO</button>
      <button class="tab-btn" :class="{ active: activeTab === 'passengers' }" @click="activeTab = 'passengers'">PASSENGERS</button>
      <button class="tab-btn" :class="{ active: activeTab === 'modules' }" @click="activeTab = 'modules'">MODULES</button>
    </div>

    <div class="tab-content">
      
      <div v-if="activeTab === 'cargo'" class="tab-pane">
        
        <div class="section-block">
            <div class="section-title">:: CARGO MANIFEST ({{ onboardCargo.length }} / {{ store.ship?.cargo_capacity }}) ::</div>
            <div v-if="onboardCargo.length === 0" class="empty-msg">HOLD EMPTY</div>
            
            <div v-for="c in onboardCargo" :key="c.id" class="compact-row onboard">
                <div class="col-main">
                    <span class="name">{{ c.item_name }} ({{ c.quantity }})</span>
                    <span class="dest">To: {{ getPlanetName(c.destination_key) }}</span>
                </div>
                <div class="col-meta">
                    <span class="pay">{{ c.payout }}cr</span>
                    <button class="btn-xs btn-drop" @click="handleDrop(c.id)">DROP</button>
                </div>
            </div>
        </div>

        <div class="section-block">
            <div class="section-title">:: LOCAL FREIGHT MARKET ::</div>
            <div v-if="marketCargo.length === 0" class="empty-msg">NO JOBS AVAILABLE</div>

            <div v-for="c in marketCargo" :key="c.id" class="compact-row market">
                <div class="col-main">
                    <span class="name">{{ c.item_name }} ({{ c.quantity }})</span>
                    <span class="dest">-> {{ getPlanetName(c.destination_key) }}</span>
                </div>
                <div class="col-meta">
                    <span class="pay">{{ c.payout }}cr</span>
                    <button class="btn-xs btn-accept" @click="handleAccept(c.id)" :disabled="store.uiState.isLoading">GET</button>
                </div>
            </div>
        </div>
      </div>

      <div v-if="activeTab === 'passengers'" class="tab-pane">
        
        <div class="section-block">
            <div class="section-title">:: PAX MANIFEST ({{ onboardPassengers.length }} / {{ store.ship?.passenger_slots }}) ::</div>
            <div v-if="onboardPassengers.length === 0" class="empty-msg">CABINS EMPTY</div>

            <div v-for="c in onboardPassengers" :key="c.id" class="compact-row onboard">
                <div class="col-main">
                    <span class="name">{{ c.item_name }} ({{ c.quantity }})</span>
                    <span class="dest">To: {{ getPlanetName(c.destination_key) }}</span>
                </div>
                <div class="col-meta">
                    <span class="pay">{{ c.payout }}cr</span>
                    <button class="btn-xs btn-drop" @click="handleDrop(c.id)">EVICT</button>
                </div>
            </div>
        </div>

        <div class="section-block">
            <div class="section-title">:: TRANSPORT REQUESTS ::</div>
            <div v-if="marketPassengers.length === 0" class="empty-msg">NO PASSENGERS WAITING</div>

            <div v-for="c in marketPassengers" :key="c.id" class="compact-row market">
                <div class="col-main">
                    <span class="name">{{ c.item_name }} ({{ c.quantity }})</span>
                    <span class="dest">-> {{ getPlanetName(c.destination_key) }}</span>
                </div>
                <div class="col-meta">
                    <span class="pay">{{ c.payout }}cr</span>
                    <button class="btn-xs btn-accept" @click="handleAccept(c.id)" :disabled="store.uiState.isLoading">BOARD</button>
                </div>
            </div>
        </div>
      </div>

      <div v-if="activeTab === 'modules'" class="tab-pane">
        <div class="section-block">
            <div class="section-title">:: INSTALLED SYSTEMS ({{ installedModules.length }} / {{ store.ship?.max_module_slots }}) ::</div>
            <div v-if="installedModules.length === 0" class="empty-msg">NO MODS INSTALLED</div>

            <div v-for="(m, i) in installedModules" :key="i" class="compact-row installed">
                <div class="col-main">
                    <span class="name">{{ m.name }}</span>
                    <span class="effect-sm">{{ m.stat_modifier }} +{{ m.stat_value }}</span>
                </div>
            </div>
        </div>

        <div class="section-block">
            <div class="section-title">:: SHIPYARD CATALOG ::</div>
            <div v-if="marketModules.length === 0" class="empty-msg">NO UPGRADES HERE</div>

            <div v-for="m in marketModules" :key="m.key" class="compact-row market">
                <div class="col-main">
                    <span class="name">{{ m.name }}</span>
                    <span class="cost-sm">{{ m.cost }}cr</span>
                </div>
                <div class="col-meta">
                     <span class="effect-sm">{{ m.stat_modifier }} {{ m.stat_value > 0 ? '+' : ''}}{{ m.stat_value }}</span>
                     <button class="btn-xs btn-buy" @click="handleBuyModule(m.key)" :disabled="store.uiState.isLoading || (store.ship?.credits || 0) < m.cost">BUY</button>
                </div>
            </div>
        </div>
      </div>

    </div>
  </div>
</template>

<style scoped>
.ops-panel {
    display: flex; flex-direction: column; height: 100%;
    font-family: 'Courier New', monospace;
    background: #000;
}

/* TABS STYLING */
.tabs-header {
    display: flex;
    border-bottom: 2px solid #004400;
    background: #001100;
}

.tab-btn {
    flex: 1; background: transparent; border: none;
    border-right: 1px solid #004400; color: #006600;
    padding: 8px 5px; cursor: pointer; font-weight: bold;
    font-family: inherit; font-size: 0.85rem;
    transition: all 0.2s;
}
.tab-btn:hover { background: #002200; color: #00ff41; }
.tab-btn.active { background: #004400; color: #fff; text-shadow: 0 0 5px #00ff41; }

.tab-content { flex: 1; overflow-y: auto; padding: 10px; }

/* SECTIONS */
.section-block { margin-bottom: 20px; }
.section-title { 
    color: #008f11; border-bottom: 1px dashed #004400; 
    margin-bottom: 8px; padding-bottom: 2px; font-size: 0.8rem;
}
.empty-msg { color: #004400; font-style: italic; text-align: center; margin: 10px 0; font-size: 0.8rem; }

/* COMPACT ROWS */
.compact-row {
    display: flex; align-items: center; justify-content: space-between;
    background: rgba(0,20,0,0.3); border-bottom: 1px solid #002200;
    padding: 4px 6px; font-size: 0.85rem;
}
.compact-row:hover { background: rgba(0,40,0,0.5); }
.compact-row.onboard { border-left: 2px solid #008f11; }
.compact-row.market { border-left: 2px solid #004400; }
.compact-row.installed { border-left: 2px solid #0066ff; }

.col-main { display: flex; align-items: center; gap: 8px; flex: 1; overflow: hidden; }
.col-meta { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }

.name { color: #fff; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-weight: bold; }
.dest { color: #ffff00; font-size: 0.8rem; }
.pay { color: #00ff41; font-weight: bold; }
.cost-sm { color: #ffaa00; font-size: 0.8rem; }
.effect-sm { color: #aaaaff; font-size: 0.75rem; font-style: italic; }

/* MINI BUTTONS */
.btn-xs { 
    font-family: inherit; font-weight: bold; cursor: pointer; border: none; 
    padding: 2px 6px; font-size: 0.75rem; 
}
.btn-accept { background: #006600; color: #fff; }
.btn-accept:hover { background: #00ff41; color: #000; }
.btn-drop { background: #440000; color: #ffaaaa; }
.btn-drop:hover { background: #ff0000; color: #fff; }
.btn-buy { background: #aa6600; color: #000; }
.btn-buy:hover { background: #ffaa00; }
.btn-buy:disabled { background: #332200; color: #664400; cursor: not-allowed; }
</style>