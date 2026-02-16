<script setup lang="ts">
import { useGameStore } from '../stores/gameStore'
import { computed } from 'vue'

const store = useGameStore()

const eventList = computed(() => store.arrivalEvents)

function dismiss() {
    store.clearEvents()
}
</script>

<template>
  <div v-if="store.uiState.showEvents" class="overlay">
    <div class="modal">
      <div class="header">
        ⚠ SHIP LOG: ARRIVAL REPORT ⚠
      </div>
      
      <div class="content">
        <div v-for="(evt, i) in eventList" :key="i" class="event-item">
            <div class="event-type">:: {{ evt.type.toUpperCase() }} DETECTED</div>
            <div class="event-desc">{{ evt.description }}</div>
            <div class="event-effect">{{ evt.effect }}</div>
        </div>
      </div>

      <div class="footer">
        <button @click="dismiss">ACKNOWLEDGE</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
    position: fixed; top: 0; left: 0; width: 100%; height: 100%;
    background: rgba(0,0,0,0.7); z-index: 2000;
    display: flex; align-items: center; justify-content: center;
}

.modal {
    background: #001100;
    border: 2px solid #ff3333; /* Alert Color */
    width: 400px;
    max-width: 90%;
    font-family: 'Courier New', monospace;
    box-shadow: 0 0 20px rgba(255, 51, 51, 0.3);
}

.header {
    background: #440000; color: #fff; padding: 10px; font-weight: bold; text-align: center;
    border-bottom: 1px solid #ff3333; letter-spacing: 1px;
}

.content {
    padding: 20px;
    max-height: 300px; overflow-y: auto;
}

.event-item {
    margin-bottom: 20px;
    border-left: 3px solid #ff3333;
    padding-left: 10px;
}

.event-type { color: #ff3333; font-weight: bold; margin-bottom: 5px; }
.event-desc { color: #fff; font-size: 0.9rem; margin-bottom: 5px; }
.event-effect { color: #ffff00; font-weight: bold; font-size: 0.9rem; }

.footer {
    padding: 10px; text-align: center; border-top: 1px solid #440000;
}

button {
    background: #ff3333; color: #fff; border: none; padding: 10px 30px;
    font-family: 'Courier New', monospace; font-weight: bold; cursor: pointer;
}
button:hover { background: #fff; color: #000; }
</style>