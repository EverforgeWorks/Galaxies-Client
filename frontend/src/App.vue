<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameStore } from './stores/gameStore'

import StarMap from './components/StarMap.vue'
import ShipStatus from './components/ShipStatus.vue'
import OperationsPanel from './components/OperationsPanel.vue'
import EventPopup from './components/EventPopup.vue'
import SaveSlotScreen from './components/SaveSlotScreen.vue'
import OnboardingScreen from './components/OnboardingScreen.vue'

const store = useGameStore()

onMounted(() => { 
    store.initGameEvents()
})
</script>

<template>
  <div class="app-container">
    <SaveSlotScreen v-if="store.currentView === 'menu'" />
    <OnboardingScreen v-else-if="store.currentView === 'onboarding'" />
    
    <div v-else-if="store.currentView === 'game'" class="terminal-grid">
        <div class="scanline"></div>
        <EventPopup />

        <div class="col-left">
          <div class="section-status">
            <ShipStatus />
          </div>
          <div class="section-ops">
            <OperationsPanel />
          </div>
          <div class="section-comms">
            <div class="panel-header">
                <span>:: SYSTEM LOG ::</span>
            </div>
            <div class="comms-content">
                <div class="log-placeholder">SLOT {{ store.activeSlot }} - UPLINK ACTIVE</div>
            </div>
          </div>
        </div>

        <div class="col-right">
          <StarMap />
        </div>
    </div>
  </div>
</template>

<style>
.app-container {
    height: 100vh;
    width: 100vw;
    overflow: hidden;
}

/* TERMINAL GRID CSS UNCHANGED FROM ORIGINAL */
.terminal-grid {
  display: grid;
  grid-template-columns: 450px 1fr;
  height: 100vh; width: 100vw;
  background: #050505; color: #00ff41;
  overflow: hidden;
}

.col-left {
  display: flex; flex-direction: column;
  border-right: 2px solid #004400;
  background: rgba(0,20,0,0.3);
  min-width: 0; height: 100%;
  overflow: hidden;
}

.col-right { position: relative; background: #000; min-width: 0; height: 100%; overflow: hidden; }
.section-status { flex-shrink: 0; border-bottom: 1px solid #004400; }
.section-ops { flex: 1; overflow-y: auto; border-bottom: 1px solid #004400; min-height: 0; }
.section-comms { height: 150px; flex-shrink: 0; display: flex; flex-direction: column; border-top: 1px solid #004400; }
.panel-header { background: #002200; color: #008f11; padding: 4px 10px; font-size: 0.75rem; letter-spacing: 2px; display: flex; justify-content: space-between; }
.comms-content { flex: 1; overflow: hidden; background: rgba(0,0,0,0.8); padding: 10px; }
.log-placeholder { color: #004400; font-family: 'Courier New', monospace; font-size: 0.8rem; text-align: center; margin-top: 20px; }
.scanline { position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: linear-gradient(to bottom, rgba(255,255,255,0), rgba(255,255,255,0) 50%, rgba(0,0,0,0.1) 50%, rgba(0,0,0,0.1)); background-size: 100% 4px; pointer-events: none; z-index: 9999; opacity: 0.3; }
</style>