<script setup>
import { ref } from 'vue'
import { nodes } from '@/composables/useData'
import { useTopo } from '@/composables/useTopo'

const fullTopoCanvas = ref(null)
const { tooltipData, tooltipStyle, zoomIn, zoomOut, reset, handleMouseMove, handleMouseLeave } = useTopo(fullTopoCanvas, {
  fullMode: true,
  headerH: 0,
})

const layerOptions = ['物理层', '逻辑层', 'OSPF 域']
const selectedLayer = ref(layerOptions[0])
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">网络拓扑图</h2>
      <div class="page-controls">
        <button class="ctrl-btn" @click="zoomIn">放大</button>
        <button class="ctrl-btn" @click="zoomOut">缩小</button>
        <button class="ctrl-btn" @click="reset">复位</button>
        <select class="ctrl-select" v-model="selectedLayer">
          <option v-for="opt in layerOptions" :key="opt">{{ opt }}</option>
        </select>
      </div>
    </div>
    <div class="topo-full-wrap">
      <canvas
        ref="fullTopoCanvas"
        id="fullTopo"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave"
      ></canvas>
      <div class="topo-legend-panel">
        <div class="tl-item"><div class="tl-dot ok"></div><span>正常</span></div>
        <div class="tl-item"><div class="tl-dot warn"></div><span>警告</span></div>
        <div class="tl-item"><div class="tl-dot err"></div><span>故障</span></div>
        <div class="tl-sep"></div>
        <div class="tl-item"><div class="tl-line solid"></div><span>光纤链路</span></div>
        <div class="tl-item"><div class="tl-line dashed"></div><span>备用链路</span></div>
      </div>
      <div class="topo-tooltip" :style="tooltipStyle" v-if="tooltipData">
        <div class="tt-title">{{ tooltipData.topoNode.label }}</div>
        <template v-if="tooltipData.realNode">
          <div class="tt-row"><span class="tt-label">IP</span><span class="tt-val">{{ tooltipData.realNode.ip }}</span></div>
          <div class="tt-row">
            <span class="tt-label">状态</span>
            <span class="tt-val" :style="{color:tooltipData.topoNode.color}">
              {{ tooltipData.realNode.status === 'ok' ? '正常' : tooltipData.realNode.status === 'warn' ? '警告' : '故障' }}
            </span>
          </div>
          <div class="tt-row"><span class="tt-label">时延</span><span class="tt-val">{{ tooltipData.realNode.lat || '—' }}ms</span></div>
          <div class="tt-row"><span class="tt-label">CPU</span><span class="tt-val">{{ tooltipData.realNode.cpu || '—' }}%</span></div>
        </template>
      </div>
    </div>
  </div>
</template>
