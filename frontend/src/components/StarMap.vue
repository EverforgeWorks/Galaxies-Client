<script setup lang="ts">
/**
 * StarMap Component (Refactored)
 * Implements "Server-First" Navigation:
 * 1. User Clicks Warp
 * 2. Client sends request to Go Backend
 * 3. Awaits Success/Fail
 * 4. IF Success: Play Animation using snapshot of start coordinates
 * 5. IF Fail: Show Error, do not move.
 */

import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useGameStore } from '../stores/gameStore'

const store = useGameStore()

// --- STATE MANAGEMENT ---
const canvasRef = ref<HTMLCanvasElement | null>(null)
const containerRef = ref<HTMLElement | null>(null)
const selectedStar = ref<any>(null)

// Animation State
const isWarping = ref(false)
const warpProgress = ref(0) // 0.0 to 1.0
let animationFrameId: number | null = null
let animationStartTime: number = 0

// Critical: Snapshot of where the ship WAS before the warp command succeeded
const animationStartCoords = ref<number[]>([0,0])

// --- CONSTANTS ---
const GAME_WORLD_SIZE = 55 
const WARP_DURATION_MS = 2500 

// --- COMPUTED DATA ---
const currentPlanetObj = computed(() => {
    return store.universe.find(p => p.key === store.ship?.location_key)
})

const flightPlan = computed(() => {
    if (!selectedStar.value || !currentPlanetObj.value || !store.ship) return null
    if (selectedStar.value.key === store.ship.location_key) return null

    const p1 = currentPlanetObj.value.coordinates || [0,0]
    const p2 = selectedStar.value.coordinates || [0,0]
    
    // Euclidean distance
    const dist = Math.sqrt(Math.pow(p2[0]-p1[0], 2) + Math.pow(p2[1]-p1[1], 2))
    const cost = Math.ceil(dist * (store.ship.base_burn_rate || 1.0))
    
    return { 
        distance: dist.toFixed(1), 
        cost: cost, 
        canAfford: store.ship.fuel >= cost 
    }
})

// --- WARP LOGIC ---

async function startWarpSequence() {
  if (isWarping.value || !flightPlan.value || !selectedStar.value) return
  if (!currentPlanetObj.value) return

  // 1. Snapshot Start Position (Critical for animation)
  animationStartCoords.value = [...currentPlanetObj.value.coordinates]

  // 2. Execute Server Call FIRST (Server-First Authority)
  const destKey = selectedStar.value.key
  
  // This will await the Wails call. 
  // If successful, store.ship.location_key updates immediately to the NEW planet.
  const success = await store.travel(destKey)

  // 3. If failed (e.g. 402 Payment Required), abort - Do not animate
  if (!success) {
      selectedStar.value = null // Deselect to clear state
      return
  }

  // 4. If success, Play Animation
  // Note: store.ship.location_key is ALREADY the new planet now.
  // We use animationStartCoords to draw the ship at the "Old" location and lerp to "New".
  isWarping.value = true
  warpProgress.value = 0
  animationStartTime = performance.now()
  animateFrame()
}

function animateFrame() {
  const now = performance.now()
  const elapsed = now - animationStartTime
  
  // Calculate normalized progress (0 to 1)
  const progress = Math.min(elapsed / WARP_DURATION_MS, 1.0)
  warpProgress.value = progress

  draw() 

  if (progress < 1.0) {
    animationFrameId = requestAnimationFrame(animateFrame)
  } else {
    // Animation Complete
    isWarping.value = false
    animationFrameId = null
    selectedStar.value = null
    
    // Trigger a full data refresh to get new market data/contracts for the new planet
    store.refreshAll()
  }
}

// --- DRAWING ENGINE ---

function draw() {
  const ctx = canvasRef.value?.getContext('2d')
  if (!ctx || !canvasRef.value) return

  // 1. Get Canvas Dimensions
  const w = canvasRef.value.width
  const h = canvasRef.value.height
  
  // Safety check
  if (w === 0 || h === 0) return 

  const cx = w / 2
  const cy = h / 2
  
  // Scale: Pixels per Light Year
  const scale = Math.min(w, h) / GAME_WORLD_SIZE

  // 2. Clear Background
  ctx.fillStyle = '#050505'
  ctx.fillRect(0, 0, w, h)

  // 3. Grid (Static 5-unit lines)
  ctx.strokeStyle = '#064e3b'
  ctx.lineWidth = 0.5
  const gridSize = 5 * scale 
  
  ctx.beginPath()
  for (let x = cx; x < w; x += gridSize) { ctx.moveTo(x, 0); ctx.lineTo(x, h); }
  for (let x = cx; x > 0; x -= gridSize) { ctx.moveTo(x, 0); ctx.lineTo(x, h); }
  for (let y = cy; y < h; y += gridSize) { ctx.moveTo(0, y); ctx.lineTo(w, y); }
  for (let y = cy; y > 0; y -= gridSize) { ctx.moveTo(0, y); ctx.lineTo(w, y); }
  ctx.stroke()

  // 4. Center Crosshair
  ctx.strokeStyle = '#004400'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(cx, cy - 10); ctx.lineTo(cx, cy + 10)
  ctx.moveTo(cx - 10, cy); ctx.lineTo(cx + 10, cy)
  ctx.stroke()

  // 5. CHECK DATA EXISTENCE
  if (!store.universe || store.universe.length === 0) {
      ctx.fillStyle = '#ff3333'
      ctx.font = '16px "Courier New"'
      ctx.textAlign = 'center'
      ctx.fillText("NO CHART DATA FOUND", cx, cy - 20)
      ctx.fillStyle = '#008f11'
      ctx.font = '12px "Courier New"'
      ctx.fillText("CHECK SENSOR UPLINK", cx, cy + 20)
      return
  }

  // 6. Draw Hyperlanes (Background layer)
  ctx.strokeStyle = '#008F11'
  ctx.lineWidth = 1
  ctx.globalAlpha = 0.15
  ctx.beginPath()
  store.universe.forEach(p1 => {
    store.universe.forEach(p2 => {
      const c1 = p1.coordinates || [0,0]
      const c2 = p2.coordinates || [0,0]
      const dist = Math.sqrt(Math.pow(c1[0]-c2[0], 2) + Math.pow(c1[1]-c2[1], 2))
      // Draw line if planets are close enough (simulating established lanes)
      if (dist < 18 && p1 !== p2) { 
         const x1 = cx + c1[0] * scale
         const y1 = cy - c1[1] * scale
         const x2 = cx + c2[0] * scale
         const y2 = cy - c2[1] * scale
         ctx.moveTo(x1, y1); ctx.lineTo(x2, y2)
      }
    })
  })
  ctx.stroke()
  ctx.globalAlpha = 1.0

  // 7. Draw Planets
  store.universe.forEach(p => {
    if (!p.coordinates) return

    const x = cx + p.coordinates[0] * scale
    const y = cy - p.coordinates[1] * scale // Flip Y for cartesian
    
    // If warping, we are "in transit", so don't highlight the current location yet
    const isCurrent = !isWarping.value && p.key === store.ship?.location_key
    const isSelected = selectedStar.value?.key === p.key

    // Planet Dot
    ctx.beginPath()
    const radius = isCurrent ? 6 : (isSelected ? 5 : 3)
    ctx.arc(x, y, radius, 0, Math.PI * 2)
    ctx.fillStyle = isCurrent ? '#fff' : (isSelected ? '#00ff41' : '#008F11')
    ctx.fill()
    
    // Selection Ring
    if (isSelected || isCurrent) {
        ctx.beginPath()
        ctx.arc(x, y, radius + 4, 0, Math.PI * 2)
        ctx.strokeStyle = isCurrent ? 'rgba(255,255,255,0.3)' : 'rgba(0,255,65,0.5)'
        ctx.stroke()
    }

    // Name Label
    ctx.font = '11px "Courier New"'
    ctx.textAlign = 'left'
    ctx.fillStyle = isCurrent ? '#fff' : (isSelected ? '#00ff41' : '#006600')
    ctx.fillText(p.name, x + 12, y + 4)
  })
  
  // 8. Flight Vector / Target Line
  if (selectedStar.value && store.ship && !isWarping.value) {
    const p = selectedStar.value
    const x = cx + p.coordinates[0] * scale
    const y = cy - p.coordinates[1] * scale
    
    const curr = store.universe.find(p => p.key === store.ship?.location_key)
    if (curr && curr.coordinates) {
       const startX = cx + curr.coordinates[0] * scale
       const startY = cy - curr.coordinates[1] * scale
       
       ctx.beginPath()
       ctx.moveTo(startX, startY)
       ctx.lineTo(x, y)
       ctx.setLineDash([4, 4])
       ctx.strokeStyle = flightPlan.value?.canAfford ? '#00ff41' : '#ff3333'
       ctx.stroke()
       ctx.setLineDash([])
    }
  }

  // 9. ANIMATION LAYER: The Ship
  if (isWarping.value && selectedStar.value) {
    // USE SNAPSHOT COORDS (Start) -> SELECTED STAR (End)
    // We do NOT use currentPlanetObj because that has already updated to the destination!
    const startCoords = animationStartCoords.value
    const endCoords = selectedStar.value.coordinates

    // Convert Game Coords to Canvas Coords
    const startX = cx + startCoords[0] * scale
    const startY = cy - startCoords[1] * scale
    const endX = cx + endCoords[0] * scale
    const endY = cy - endCoords[1] * scale

    // Linear Interpolation (Lerp) based on warpProgress
    const currentX = startX + (endX - startX) * warpProgress.value
    const currentY = startY + (endY - startY) * warpProgress.value

    // Calculate rotation angle (atan2(dy, dx))
    const angle = Math.atan2(endY - startY, endX - startX)

    ctx.save()
    ctx.translate(currentX, currentY)
    ctx.rotate(angle) // Rotate canvas to align with flight path

    // Draw Asteroids-style Vector Ship
    ctx.beginPath()
    ctx.strokeStyle = '#ffffff'
    ctx.fillStyle = '#000000'
    ctx.lineWidth = 2
    
    // Triangle pointing right (since 0 radians is East)
    ctx.moveTo(10, 0)   // Nose
    ctx.lineTo(-8, 6)   // Back Left
    ctx.lineTo(-4, 0)   // Engine indented
    ctx.lineTo(-8, -6)  // Back Right
    ctx.closePath()
    
    ctx.fill()
    ctx.stroke()

    // Engine thruster (flickering effect)
    if (Math.random() > 0.3) {
        ctx.beginPath()
        ctx.strokeStyle = '#00ff41'
        ctx.moveTo(-6, 0)
        ctx.lineTo(-12 - (Math.random() * 5), 0)
        ctx.stroke()
    }

    ctx.restore()
  }
}

// --- INTERACTION ---
function handleClick(e: MouseEvent) {
  // Disable selection during warp animation
  if (isWarping.value) return 
  if (!canvasRef.value) return
  
  const rect = canvasRef.value.getBoundingClientRect()
  
  // Recalc Scale logic for hit detection
  const w = canvasRef.value.width
  const h = canvasRef.value.height
  const cx = w / 2
  const cy = h / 2
  const scale = Math.min(w, h) / GAME_WORLD_SIZE
  
  const mouseX = e.clientX - rect.left
  const mouseY = e.clientY - rect.top

  const clicked = store.universe?.find(p => {
    if (!p.coordinates) return false
    const px = cx + p.coordinates[0] * scale
    const py = cy - p.coordinates[1] * scale
    const dist = Math.sqrt(Math.pow(mouseX - px, 2) + Math.pow(mouseY - py, 2))
    return dist < 15
  })
  
  selectedStar.value = clicked || null
  draw()
}

// --- RESIZE LOGIC ---
let resizeObserver: ResizeObserver | null = null

onMounted(() => {
  if (containerRef.value && canvasRef.value) {
    resizeObserver = new ResizeObserver(() => {
      if (containerRef.value && canvasRef.value) {
        // Force integer pixels for sharpness
        canvasRef.value.width = Math.floor(containerRef.value.clientWidth)
        canvasRef.value.height = Math.floor(containerRef.value.clientHeight)
        draw()
      }
    })
    resizeObserver.observe(containerRef.value)
  }
  // Initial draw attempt
  setTimeout(draw, 100)
})

onUnmounted(() => {
  if (resizeObserver) resizeObserver.disconnect()
  if (animationFrameId) cancelAnimationFrame(animationFrameId)
})

// Deep watch on critical data to trigger redraws
watch(() => [store.universe, store.ship?.location_key, flightPlan.value], draw, { deep: true })
</script>

<template>
  <div ref="containerRef" class="map-container">
    <canvas ref="canvasRef" @click="handleClick"></canvas>
    
    <div v-if="selectedStar && selectedStar.key !== store.ship?.location_key && !isWarping" class="warp-controls">
      <h3>{{ selectedStar.name }}</h3>

      <div v-if="store.uiState.lastError" class="error-msg">
         ⚠ {{ store.uiState.lastError }}
      </div>

      <div v-if="flightPlan" class="trip-stats">
        <div class="stat-line"><span>DIST:</span><span>{{ flightPlan.distance }} LY</span></div>
        <div class="stat-line"><span>FUEL:</span><span :class="{ 'alert': !flightPlan.canAfford }">{{ flightPlan.cost }} UNITS</span></div>
      </div>
      
      <button 
        @click="startWarpSequence" 
        class="btn-warp" 
        :disabled="!flightPlan?.canAfford || store.uiState.isLoading" 
        :class="{ 'disabled': !flightPlan?.canAfford || store.uiState.isLoading }"
      >
        <span v-if="store.uiState.isLoading">CALCULATING...</span>
        <span v-else>{{ flightPlan?.canAfford ? 'INITIATE WARP' : 'INSUFFICIENT FUEL' }}</span>
      </button>
    </div>

    <div v-if="isWarping" class="warp-status">
        TRAJECTORY LOCKED... WARPING
    </div>
  </div>
</template>

<style scoped>
.map-container { 
    width: 100%; 
    height: 100%; 
    position: relative; 
    background: #050505; 
    overflow: hidden; 
}

canvas {
    display: block; 
    width: 100%;
    height: 100%;
}

.warp-controls {
  position: absolute; bottom: 30px; left: 50%; transform: translateX(-50%);
  background: rgba(0, 15, 0, 0.95); 
  border: 1px solid #00ff41; padding: 15px; text-align: center;
  min-width: 280px; box-shadow: 0 0 20px rgba(0, 255, 65, 0.15);
  font-family: 'Courier New', monospace; z-index: 100;
}

.error-msg {
    color: #ff3333;
    font-size: 0.75rem;
    border: 1px solid #ff3333;
    background: rgba(50,0,0,0.5);
    padding: 5px;
    margin-bottom: 10px;
    font-weight: bold;
}

.warp-status {
    position: absolute; bottom: 30px; left: 50%; transform: translateX(-50%);
    color: #00ff41; font-family: 'Courier New', monospace; 
    font-weight: bold; animation: blink 1s infinite;
    text-shadow: 0 0 10px #00ff41;
}

@keyframes blink { 50% { opacity: 0.5; } }

.warp-controls h3 { color: #00ff41; margin: 0 0 10px 0; font-size: 1.2rem; }
.trip-stats { margin-bottom: 15px; border-top: 1px dashed #004400; border-bottom: 1px dashed #004400; padding: 10px 0; }
.stat-line { display: flex; justify-content: space-between; font-size: 0.9rem; color: #00ff41; margin-bottom: 5px; }
.alert { color: #ff3333; font-weight: bold; }
.btn-warp { background: #00ff41; color: #000; border: none; padding: 12px 20px; font-weight: bold; cursor: pointer; width: 100%; font-family: 'Courier New', monospace; }
.btn-warp:hover { background: #fff; }
.btn-warp.disabled { background: #222; color: #666; cursor: not-allowed; }
</style>