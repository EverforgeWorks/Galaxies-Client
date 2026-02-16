<script setup lang="ts">
import { ref } from 'vue'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()
const playerName = ref('')
const shipName = ref('')
const selectedShipType = ref('ship_scout') // Default to one from universe.yaml

const ships = [
    { key: 'ship_scout', name: 'Pathfinder Class', desc: 'Fast, light, limited cargo.' },
    { key: 'ship_hauler', name: 'Mule Class', desc: 'Heavy, slow, high capacity.' }
]

function handleSubmit() {
    if (!playerName.value || !shipName.value) return
    if (store.activeSlot !== null) {
        store.startNewSession(store.activeSlot, playerName.value, shipName.value, selectedShipType.value)
    }
}
</script>

<template>
  <div class="menu-overlay">
    <div class="onboarding-box">
      <h2>:: COMMISSIONING LOG ::</h2>
      
      <div class="form-group">
        <label>CAPTAIN NAME</label>
        <input v-model="playerName" placeholder="Enter Moniker..." maxlength="20" />
      </div>

      <div class="form-group">
        <label>VESSEL IDENTIFIER</label>
        <input v-model="shipName" placeholder="Enter Ship Name..." maxlength="20" />
      </div>

      <div class="ship-selection">
        <label>SELECT HULL CONFIGURATION</label>
        <div class="ship-grid">
            <div 
                v-for="s in ships" :key="s.key" 
                class="ship-card" 
                :class="{ active: selectedShipType === s.key }"
                @click="selectedShipType = s.key"
            >
                <div class="hull-name">{{ s.name }}</div>
                <div class="hull-desc">{{ s.desc }}</div>
            </div>
        </div>
      </div>

      <button class="btn-finalize" @click="handleSubmit" :disabled="!playerName || !shipName">
        INITIALIZE SYSTEMS
      </button>
    </div>
  </div>
</template>

<style scoped>
.menu-overlay {
  height: 100vh; width: 100vw;
  background: #000;
  display: flex; align-items: center; justify-content: center;
  font-family: 'Courier New', monospace;
}

.onboarding-box {
  width: 450px; background: #001100; border: 1px solid #00ff41;
  padding: 30px; box-shadow: 0 0 20px rgba(0, 255, 65, 0.1);
}

h2 { color: #00ff41; font-size: 1rem; margin-bottom: 25px; text-align: center; }

.form-group { margin-bottom: 20px; display: flex; flex-direction: column; gap: 8px; }
label { color: #008f11; font-size: 0.75rem; font-weight: bold; letter-spacing: 1px; }

input {
    background: #000; border: 1px solid #004400; color: #fff;
    padding: 10px; font-family: inherit; font-size: 1rem; outline: none;
}
input:focus { border-color: #00ff41; }

.ship-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-top: 10px; }
.ship-card {
    border: 1px solid #004400; padding: 10px; cursor: pointer;
    background: rgba(0, 44, 0, 0.2);
}
.ship-card.active { border-color: #00ff41; background: rgba(0, 255, 65, 0.1); }
.hull-name { color: #fff; font-size: 0.85rem; font-weight: bold; }
.hull-desc { color: #008f11; font-size: 0.7rem; margin-top: 5px; }

.btn-finalize {
    margin-top: 30px; width: 100%; padding: 15px;
    background: #00ff41; color: #000; border: none;
    font-weight: bold; font-family: inherit; cursor: pointer;
}
.btn-finalize:disabled { background: #004400; color: #006600; cursor: not-allowed; }
</style>