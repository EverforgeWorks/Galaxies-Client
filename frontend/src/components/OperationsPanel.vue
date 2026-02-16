<script setup lang="ts">
/**
 * OpsPanel Component
 */
import { ref, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()

// --- UI STATE ---
const mode = ref('SHIP') 

const open = ref({
    shipCargo: true,
    shipPax: true,
    shipMods: false,
    planetMarket: true,
    planetJobs: true,
    planetShop: false
})

const toggle = (key: keyof typeof open.value) => { 
    open.value[key] = !open.value[key] 
}

// --- HELPERS ---

const getPlanetName = (key: string) => {
    if (!store.universe) return key
    const p = store.universe.find(x => x.key === key)
    return p ? p.name : key
}

const sortContracts = (a: any, b: any) => {
    const destA = getPlanetName(a.destination_key)
    const destB = getPlanetName(b.destination_key)
    const nameCompare = destA.localeCompare(destB)
    if (nameCompare !== 0) return nameCompare
    return (b.payout || 0) - (a.payout || 0)
}

// --- COMPUTED DATA ---

// 1. SHIP: Cargo
const shipCargo = computed(() => {
    const list = store.ship?.active_contracts?.filter((c:any) => c.type === 'cargo') || []
    return [...list].sort(sortContracts)
})
// New: Calculate total volume
const cargoVolume = computed(() => shipCargo.value.reduce((sum, c) => sum + c.quantity, 0))

// 2. SHIP: Passengers
const shipPax = computed(() => {
    const list = store.ship?.active_contracts?.filter((c:any) => c.type === 'passenger') || []
    return [...list].sort(sortContracts)
})
// New: Calculate total pax
const paxCount = computed(() => shipPax.value.reduce((sum, c) => sum + c.quantity, 0))

// 3. SHIP: Modules
const shipMods = computed(() => store.ship?.installed_modules || [])

// 4. PLANET: Market
const planetMarket = computed(() => {
    const list = store.availableJobs?.filter((j:any) => j.type === 'cargo') || []
    return [...list].sort(sortContracts)
})

// 5. PLANET: Jobs
const planetJobs = computed(() => {
    const list = store.availableJobs?.filter((j:any) => j.type === 'passenger') || []
    return [...list].sort(sortContracts)
})

// 6. PLANET: Shop
const planetShop = computed(() => store.availableModules || [])

</script>

<template>
  <div class="ops-panel">
    
    <div v-if="store.uiState.lastError" class="global-error" @click="store.uiState.lastError = null">
        ⚠ ERROR: {{ store.uiState.lastError }}
    </div>

    <div class="tabs">
        <button :class="{ active: mode === 'SHIP' }" @click="mode = 'SHIP'">SHIP MANIFEST</button>
        <button :class="{ active: mode === 'PLANET' }" @click="mode = 'PLANET'">PLANET SERVICES</button>
    </div>

    <div class="scroll-content">
        
        <div v-if="mode === 'SHIP'">
            
            <div class="section-header" @click="toggle('shipCargo')">
                <span>:: CARGO HOLD ({{ cargoVolume }} / {{ store.ship?.cargo_capacity }})</span>
                <span>{{ open.shipCargo ? '[-]' : '[+]' }}</span>
            </div>
            <div v-if="open.shipCargo" class="list-group">
                <div v-for="item in shipCargo" :key="item.id" class="list-item">
                    <div class="col-main">
                        <span class="name">{{ item.item_name }} ({{ item.quantity }})</span>
                        <span class="meta-sub">Val: {{ item.payout }}cr</span>
                    </div>
                    <span class="dest">-> {{ getPlanetName(item.destination_key) }}</span>
                    <button class="btn-xs warn" @click="store.dropContract(item.id)">DUMP</button>
                </div>
                <div v-if="!shipCargo.length" class="empty">-- EMPTY --</div>
            </div>

            <div class="section-header" @click="toggle('shipPax')">
                <span>:: CABINS ({{ paxCount }} / {{ store.ship?.passenger_slots }})</span>
                <span>{{ open.shipPax ? '[-]' : '[+]' }}</span>
            </div>
            <div v-if="open.shipPax" class="list-group">
                <div v-for="pax in shipPax" :key="pax.id" class="list-item">
                    <div class="col-main">
                        <span class="name">PASSENGER</span>
                        <span class="meta-sub">Fare: {{ pax.payout }}cr</span>
                    </div>
                    <span class="dest">-> {{ getPlanetName(pax.destination_key) }}</span>
                    <button class="btn-xs warn" @click="store.dropContract(pax.id)">EJECT</button>
                </div>
                <div v-if="!shipPax.length" class="empty">-- EMPTY --</div>
            </div>

            <div class="section-header" @click="toggle('shipMods')">
                <span>:: SYSTEMS ({{ shipMods.length }})</span>
                <span>{{ open.shipMods ? '[-]' : '[+]' }}</span>
            </div>
            <div v-if="open.shipMods" class="list-group">
                <div v-for="(mod, i) in shipMods" :key="i" class="list-item">
                    <span class="name white">{{ mod.name }}</span>
                    <span class="meta">{{ mod.stat_modifier }} +{{ mod.stat_value }}</span>
                </div>
                <div v-if="!shipMods.length" class="empty">-- STOCK CONFIG --</div>
            </div>
        </div>

        <div v-if="mode === 'PLANET'">
            <div class="section-header" @click="toggle('planetMarket')">
                <span>:: LOCAL EXPORTS ({{ planetMarket.length }})</span>
                <span>{{ open.planetMarket ? '[-]' : '[+]' }}</span>
            </div>
            <div v-if="open.planetMarket" class="list-group">
                <div v-for="job in planetMarket" :key="job.id" class="list-item">
                    <div class="col-main">
                        <span class="name">{{ job.item_name }} ({{ job.quantity }})</span>
                    </div>
                    <span class="dest">-> {{ getPlanetName(job.destination_key) }}</span>
                    <span class="pay">{{ job.payout }}cr</span>
                    <button class="btn-xs" @click="store.acceptContract(job.id)">TAKE</button>
                </div>
                <div v-if="!planetMarket.length" class="empty">-- NO CONTRACTS --</div>
            </div>

            <div class="section-header" @click="toggle('planetJobs')">
                <span>:: TRANSPORT REQUESTS ({{ planetJobs.length }})</span>
                <span>{{ open.planetJobs ? '[-]' : '[+]' }}</span>
            </div>
            <div v-if="open.planetJobs" class="list-group">
                <div v-for="job in planetJobs" :key="job.id" class="list-item">
                    <div class="col-main">
                        <span class="name">PASSENGER</span>
                    </div>
                    <span class="dest">-> {{ getPlanetName(job.destination_key) }}</span>
                    <span class="pay">{{ job.payout }}cr</span>
                    <button class="btn-xs" @click="store.acceptContract(job.id)">BOARD</button>
                </div>
                <div v-if="!planetJobs.length" class="empty">-- NO PASSENGERS --</div>
            </div>

            <div v-if="planetShop.length > 0">
                <div class="section-header" @click="toggle('planetShop')">
                    <span>:: OUTFITTING</span>
                    <span>{{ open.planetShop ? '[-]' : '[+]' }}</span>
                </div>
                <div v-if="open.planetShop" class="list-group">
                    <div v-for="mod in planetShop" :key="mod.key" class="list-item">
                        <span class="name white">{{ mod.name }}</span>
                        <span class="pay">{{ mod.cost }}cr</span>
                        <button class="btn-xs" :disabled="(store.ship?.credits || 0) < mod.cost" @click="store.buyModule(mod.key)">BUY</button>
                    </div>
                </div>
            </div>

        </div>

    </div>
  </div>
</template>

<style scoped>
/* Main Container */
.ops-panel { display: flex; flex-direction: column; height: 100%; font-family: 'Courier New', monospace; }

/* Global Error */
.global-error {
    background: #440000;
    color: #ff5555;
    border-bottom: 1px solid #ff0000;
    padding: 8px;
    text-align: center;
    font-size: 0.8rem;
    cursor: pointer;
    font-weight: bold;
}

/* Tabs */
.tabs { display: flex; background: #001100; border-bottom: 1px solid #004400; flex-shrink: 0; }
.tabs button {
    flex: 1; background: transparent; border: none; color: #006600; 
    padding: 8px; font-weight: bold; cursor: pointer; border-right: 1px solid #002200;
    transition: all 0.2s; font-size: 0.8rem;
}
.tabs button.active { color: #000; background: #00ff41; }

/* Content Area */
.scroll-content { flex: 1; overflow-y: auto; padding-bottom: 20px; }

/* Accordion Headers */
.section-header {
    background: #002200; color: #00ff41; padding: 5px 10px; font-size: 0.75rem;
    border-bottom: 1px solid #004400; border-top: 1px solid #004400;
    display: flex; justify-content: space-between; cursor: pointer; margin-top: 5px;
}
.section-header:hover { background: #003300; }

/* List Items */
.list-group { background: rgba(0,20,0,0.2); }
.list-item {
    display: flex; align-items: center; justify-content: space-between;
    padding: 4px 10px; border-bottom: 1px solid rgba(0,68,0,0.3); font-size: 0.8rem;
}
.list-item:hover { background: rgba(0,255,65,0.05); }

/* Column Layout for Items */
.col-main { display: flex; flex-direction: column; flex: 2; margin-right: 5px; overflow: hidden; }

/* Text Styles */
.name { color: #00aa00; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.name.white { color: #ccc; }
.dest { color: #006600; flex: 1; font-size: 0.7rem; text-align: left; }
.pay { color: #fff; flex: 0 0 60px; text-align: right; margin-right: 10px; font-size: 0.75rem; }
.meta { color: #006600; font-size: 0.7rem; }
.meta-sub { color: #005500; font-size: 0.65rem; }

/* Buttons */
.btn-xs {
    background: #004400; color: #00ff41; border: 1px solid #008f11;
    font-size: 0.65rem; padding: 1px 6px; cursor: pointer;
}
.btn-xs:hover { background: #00ff41; color: #000; }
.btn-xs.warn:hover { background: #ff3333; color: #fff; }
.btn-xs:disabled { opacity: 0.3; cursor: not-allowed; }

.empty { font-style: italic; color: #004400; font-size: 0.7rem; padding: 5px 10px; }
</style>