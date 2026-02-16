<script setup lang="ts">
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()
const slots = [1, 2, 3]

function handleSlotSelect(slot: number) {
    store.activeSlot = slot
    // In a real app, we would check if a file exists here via a Go helper.
    // For now, we allow the user to choose to Load or Create.
}

function startNew(slot: number) {
    store.activeSlot = slot
    store.currentView = 'onboarding'
}

function loadExisting(slot: number) {
    store.loadSession(slot)
}
</script>

<template>
  <div class="menu-overlay">
    <div class="menu-box">
      <h1 class="glitch" data-text="GALAXIES: BURN RATE">GALAXIES: BURN RATE</h1>
      <p class="subtitle">SELECT DATA UPLINK SLOT</p>

      <div class="slot-list">
        <div v-for="slot in slots" :key="slot" class="slot-card">
          <div class="slot-info">
            <span class="slot-num">SLOT 0{{ slot }}</span>
            <span class="slot-status">READY FOR SYNC</span>
          </div>
          <div class="slot-actions">
            <button class="btn-menu" @click="loadExisting(slot)">LOAD DATA</button>
            <button class="btn-menu secondary" @click="startNew(slot)">NEW JOURNEY</button>
          </div>
        </div>
      </div>
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

.menu-box {
  width: 500px; text-align: center;
  border: 1px solid #004400; padding: 40px;
  background: rgba(0, 20, 0, 0.5);
}

h1 { color: #00ff41; font-size: 2rem; margin-bottom: 5px; letter-spacing: 4px; }
.subtitle { color: #008f11; font-size: 0.8rem; margin-bottom: 30px; }

.slot-list { display: flex; flex-direction: column; gap: 15px; }

.slot-card {
  border: 1px solid #004400; background: #001100;
  padding: 15px; display: flex; justify-content: space-between; align-items: center;
}

.slot-info { display: flex; flex-direction: column; text-align: left; }
.slot-num { color: #fff; font-weight: bold; }
.slot-status { color: #006600; font-size: 0.7rem; }

.slot-actions { display: flex; gap: 10px; }

.btn-menu {
  background: #00ff41; color: #000; border: none;
  padding: 8px 15px; font-weight: bold; cursor: pointer;
  font-family: inherit; font-size: 0.75rem;
}
.btn-menu.secondary { background: #004400; color: #00ff41; }
.btn-menu:hover { background: #fff; color: #000; }
</style>