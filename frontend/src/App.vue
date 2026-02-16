<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameStore } from './stores/gameStore' // Converted to use store

import StarMap from './components/StarMap.vue'
import ShipStatus from './components/ShipStatus.vue'
import OperationsPanel from './components/OperationsPanel.vue'
import ChatLog from './components/ChatLog.vue'
import { ref } from 'vue'

const store = useGameStore()
const chatCollapsed = ref(false)

// Bootstrapping
onMounted(() => { 
    store.refreshAll()
    store.connectSocket()
})
</script>

<template>
  <div class="terminal-grid">
    <div class="scanline"></div>

    <div class="col-left">
      
      <div class="section-status">
        <ShipStatus />
      </div>

      <div class="section-ops">
        <OperationsPanel />
      </div>

      <div class="section-comms" :class="{ 'collapsed': chatCollapsed }">
        <div class="panel-header" @click="chatCollapsed = !chatCollapsed">
          <span>:: COMMS ::</span>
          <span class="toggle-icon">{{ chatCollapsed ? '▲' : '▼' }}</span>
        </div>
        <div class="comms-content" v-show="!chatCollapsed">
          <ChatLog />
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
  height: 200px; /* Default Height */
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  transition: height 0.3s ease;
  border-top: 1px solid #004400;
}

.section-comms.collapsed {
  height: 30px; /* Header only */
}

.panel-header {
  background: #002200;
  color: #008f11;
  padding: 4px 10px;
  font-size: 0.75rem;
  letter-spacing: 2px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
}
.panel-header:hover { background: #003300; color: #fff; }

.comms-content {
  flex: 1;
  overflow: hidden;
  background: rgba(0,0,0,0.8);
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