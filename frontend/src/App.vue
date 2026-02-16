<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useGameStore } from './stores/gameStore'

import StarMap from './components/StarMap.vue'
import ShipStatus from './components/ShipStatus.vue'
import OperationsPanel from './components/OperationsPanel.vue'
import EventPopup from './components/EventPopup.vue'

const store = useGameStore()

// Bootstrapping
onMounted(() => { 
    // Load initial data
    store.refreshAll()
    // Initialize Wails Event Listeners (replacing WebSockets)
    // Ensure you have updated gameStore.ts to contain this method
    if (store.initGameEvents) {
        store.initGameEvents()
    }
})
</script>

<template>
  <div class="terminal-grid">
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
            <div class="log-placeholder">OFFLINE MODE ACTIVE</div>
        </div>
      </div>

    </div>

    <div class="col-right">
      <StarMap />
    </div>

  </div>
</template>

<style>
/* FIXED LAYOUT - Prevents "Planets Disappearing" */
.terminal-grid {
  display: grid;
  grid-template-columns: 450px 1fr; /* Fixed left, Flex right */
  height: 100vh;
  width: 100vw;
  background: #050505;
  color: #00ff41;
  overflow: hidden; /* Critical */
}

.col-left {
  display: flex;
  flex-direction: column;
  border-right: 2px solid #004400;
  background: rgba(0,20,0,0.3);
  min-width: 0; /* Critical for Flexbox nesting */
  height: 100%;
  overflow: hidden;
}

.col-right {
  position: relative;
  background: #000;
  min-width: 0;
  height: 100%;
  overflow: hidden;
}

/* SECTIONS */
.section-status {
  flex-shrink: 0;
  border-bottom: 1px solid #004400;
}

.section-ops {
  flex: 1; /* Takes all remaining height */
  overflow-y: auto; /* SCROLLBAR FIX */
  border-bottom: 1px solid #004400;
  min-height: 0; /* Firefox fix */
}

.section-comms {
  height: 150px; /* Reduced Height since it's just a log now */
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-top: 1px solid #004400;
}

.panel-header {
  background: #002200;
  color: #008f11;
  padding: 4px 10px;
  font-size: 0.75rem;
  letter-spacing: 2px;
  display: flex;
  justify-content: space-between;
}

.comms-content {
  flex: 1;
  overflow: hidden;
  background: rgba(0,0,0,0.8);
  padding: 10px;
}

.log-placeholder {
    color: #004400;
    font-family: 'Courier New', monospace;
    font-size: 0.8rem;
    text-align: center;
    margin-top: 20px;
}

.scanline {
  position: fixed; top: 0; left: 0; width: 100%; height: 100%;
  background: linear-gradient(to bottom, rgba(255,255,255,0), rgba(255,255,255,0) 50%, rgba(0,0,0,0.1) 50%, rgba(0,0,0,0.1));
  background-size: 100% 4px; pointer-events: none; z-index: 9999; opacity: 0.3;
}

/* SCROLLBARS */
::-webkit-scrollbar { width: 5px; }
::-webkit-scrollbar-track { background: #001100; }
::-webkit-scrollbar-thumb { background: #004400; }
::-webkit-scrollbar-thumb:hover { background: #008f11; }
</style>