<script setup>
import { ref, onMounted, computed } from 'vue'
import { apiFetch } from '../utils/api.js'

const settingsNav = ['阈值配置', '机房配置']
const activeNav = ref(settingsNav[0])

// 机房配置数据
const idcList = ref([])
const idcSaving = ref(false)
const idcSaved = ref(false)

// 阈值表单数据
const latencyWarn = ref(3)
const latencyAlarm = ref(10)
const lostWarn = ref(0.01)
const lostAlarm = ref(0.1)
const latencyBetweenWarn = ref(15)
const latencyBetweenAlarm = ref(100)
const lostBetweenWarn = ref(5)
const lostBetweenAlarm = ref(10)
const saving = ref(false)
const saved = ref(false)

async function loadThresholds() {
  try {
    const res = await apiFetch('/api/dashboard-monitor/thresholds')
    const data = await res.json()
    if (data.latency_levels_for_idc_inner) {
      latencyWarn.value = data.latency_levels_for_idc_inner.warn
      latencyAlarm.value = data.latency_levels_for_idc_inner.alarm
    }
    if (data.lost_levels_for_idc_inner) {
      lostWarn.value = data.lost_levels_for_idc_inner.warn
      lostAlarm.value = data.lost_levels_for_idc_inner.alarm
    }
    if (data.latency_levels_for_idc_between) {
      latencyBetweenWarn.value = data.latency_levels_for_idc_between.warn
      latencyBetweenAlarm.value = data.latency_levels_for_idc_between.alarm
    }
    if (data.lost_levels_for_idc_between) {
      lostBetweenWarn.value = data.lost_levels_for_idc_between.warn
      lostBetweenAlarm.value = data.lost_levels_for_idc_between.alarm
    }
  } catch (e) {
    console.error('加载阈值配置失败:', e)
  }
}

async function saveThresholds() {
  saving.value = true
  saved.value = false
  try {
    const res = await apiFetch('/api/dashboard-monitor/thresholds', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        latency_levels_for_idc_inner:   { warn: latencyWarn.value,   alarm: latencyAlarm.value },
        lost_levels_for_idc_inner:      { warn: lostWarn.value,      alarm: lostAlarm.value },
        latency_levels_for_idc_between: { warn: latencyBetweenWarn.value, alarm: latencyBetweenAlarm.value },
        lost_levels_for_idc_between:    { warn: lostBetweenWarn.value,    alarm: lostBetweenAlarm.value }
      })
    })
    const data = await res.json()
    if (data.status === 'ok') {
      saved.value = true
      setTimeout(() => { saved.value = false }, 3000)
    }
  } catch (e) {
    console.error('保存阈值配置失败:', e)
  } finally {
    saving.value = false
  }
}

async function saveIDCs() {
  idcSaving.value = true
  idcSaved.value = false
  try {
    const res = await apiFetch('/api/idcs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(idcList.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      idcSaved.value = true
      setTimeout(() => { idcSaved.value = false }, 3000)
    }
  } catch (e) {
    console.error('保存机房配置失败:', e)
  } finally {
    idcSaving.value = false
  }
}

async function loadIDCs() {
  try {
    const res = await apiFetch('/api/idcs')
    const data = await res.json()
    if (Array.isArray(data)) idcList.value = data
  } catch (e) {
    console.error('加载机房配置失败:', e)
  }
}


onMounted(() => {
  loadThresholds()
  loadIDCs()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">系统设置</h2>
    </div>
    <div class="settings-layout">
      <div class="settings-nav">
        <div
          v-for="item in settingsNav" :key="item"
          class="sn-item" :class="{active:activeNav===item}"
          @click="activeNav = item"
        >{{ item }}</div>
      </div>
      <div class="settings-content">

        <!-- 阈值配置 Tab -->
        <div v-if="activeNav==='阈值配置'" class="settings-section">
          <div class="ts-topbar">
            <span class="ts-topbar-title">阈值配置</span>
            <div class="ts-topbar-actions">
              <button class="ctrl-btn accent-btn" :disabled="saving" @click="saveThresholds">
                {{ saving ? '保存中...' : '保存配置' }}
              </button>
              <span v-if="saved" class="ss-saved">保存成功</span>
            </div>
          </div>
          <!-- 机房内 -->
          <div class="ts-group">
            <div class="ts-group-label">机房内</div>
            <div class="ts-group-inner">
            <p class="ts-group-desc">同机房内 Leaf 节点之间的网络质量判定阈值。</p>

            <div class="ts-subgroup">
              <div class="ts-subgroup-title">时延</div>
              <p class="ts-subgroup-desc">单位：ms。低于 <span class="ts-tag ok">Warn</span> 为正常（绿色），介于 <span class="ts-tag warn">Warn</span> ~ <span class="ts-tag err">Alarm</span> 为告警（橙色），高于 <span class="ts-tag err">Alarm</span> 为严重（红色）。</p>
              <div class="ss-grid">
                <div class="ss-field">
                  <label>Warn 阈值</label>
                  <input class="ctrl-input" v-model.number="latencyWarn" type="number" step="0.1" min="0" />
                  <span class="ss-unit">ms</span>
                </div>
                <div class="ss-field">
                  <label>Alarm 阈值</label>
                  <input class="ctrl-input" v-model.number="latencyAlarm" type="number" step="0.1" min="0" />
                  <span class="ss-unit">ms</span>
                </div>
              </div>
            </div>

            <div class="ts-subgroup">
              <div class="ts-subgroup-title">丢包率</div>
              <p class="ts-subgroup-desc">单位：%。低于 <span class="ts-tag ok">Warn</span> 为正常（绿色），介于 <span class="ts-tag warn">Warn</span> ~ <span class="ts-tag err">Alarm</span> 为告警（橙色），高于 <span class="ts-tag err">Alarm</span> 为严重（红色）。</p>
              <div class="ss-grid">
                <div class="ss-field">
                  <label>Warn 阈值</label>
                  <input class="ctrl-input" v-model.number="lostWarn" type="number" step="0.001" min="0" />
                  <span class="ss-unit">%</span>
                </div>
                <div class="ss-field">
                  <label>Alarm 阈值</label>
                  <input class="ctrl-input" v-model.number="lostAlarm" type="number" step="0.001" min="0" />
                  <span class="ss-unit">%</span>
                </div>
              </div>
            </div>
            </div>
          </div>

          <!-- 机房间 -->
          <div class="ts-group">
            <div class="ts-group-label">机房间</div>
            <div class="ts-group-inner">
            <p class="ts-group-desc">不同机房之间的网络质量判定阈值，用于 SVG 拓扑图中机房节点状态展示。</p>

            <div class="ts-subgroup">
              <div class="ts-subgroup-title">时延</div>
              <p class="ts-subgroup-desc">单位：ms。低于 <span class="ts-tag ok">Warn</span> 为正常（绿色），介于 <span class="ts-tag warn">Warn</span> ~ <span class="ts-tag err">Alarm</span> 为告警（橙色），高于 <span class="ts-tag err">Alarm</span> 为严重（红色）。</p>
              <div class="ss-grid">
                <div class="ss-field">
                  <label>Warn 阈值</label>
                  <input class="ctrl-input" v-model.number="latencyBetweenWarn" type="number" step="0.1" min="0" />
                  <span class="ss-unit">ms</span>
                </div>
                <div class="ss-field">
                  <label>Alarm 阈值</label>
                  <input class="ctrl-input" v-model.number="latencyBetweenAlarm" type="number" step="0.1" min="0" />
                  <span class="ss-unit">ms</span>
                </div>
              </div>
            </div>

            <div class="ts-subgroup">
              <div class="ts-subgroup-title">丢包率</div>
              <p class="ts-subgroup-desc">单位：%。低于 <span class="ts-tag ok">Warn</span> 为正常（绿色），介于 <span class="ts-tag warn">Warn</span> ~ <span class="ts-tag err">Alarm</span> 为告警（橙色），高于 <span class="ts-tag err">Alarm</span> 为严重（红色）。</p>
              <div class="ss-grid">
                <div class="ss-field">
                  <label>Warn 阈值</label>
                  <input class="ctrl-input" v-model.number="lostBetweenWarn" type="number" step="0.001" min="0" />
                  <span class="ss-unit">%</span>
                </div>
                <div class="ss-field">
                  <label>Alarm 阈值</label>
                  <input class="ctrl-input" v-model.number="lostBetweenAlarm" type="number" step="0.001" min="0" />
                  <span class="ss-unit">%</span>
                </div>
              </div>
            </div>
            </div>
          </div>

        </div>

        <!-- 机房配置 Tab -->
        <div v-if="activeNav==='机房配置'" class="settings-section">
          <div class="ts-topbar">
            <span class="ts-topbar-title">机房配置</span>
            <div class="ts-topbar-actions">
              <button class="ctrl-btn accent-btn" :disabled="idcSaving" @click="saveIDCs">
                {{ idcSaving ? '保存中...' : '保存配置' }}
              </button>
              <span v-if="idcSaved" class="ss-saved">保存成功</span>
            </div>
          </div>
          <div class="ts-group">
            <div class="idc-table">
              <div class="idc-thead">
                <span class="idc-th code">机房编码</span>
                <span class="idc-th name">中文名称</span>
                <span class="idc-th enable">启用</span>
              </div>
              <div
                v-for="idc in idcList"
                :key="idc.code"
                class="idc-tr"
              >
                <span class="idc-td code">{{ idc.code }}</span>
                <span class="idc-td name">
                  <input class="idc-input" v-model="idc.name" type="text" />
                </span>
                <span class="idc-td enable">
                  <label class="idc-toggle">
                    <input type="checkbox" v-model="idc.enable" />
                    <span class="idc-toggle-knob"></span>
                  </label>
                </span>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
/* 阈值页顶部栏 */
.ts-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.ts-topbar-title {
  font-family: 'Orbitron', monospace;
  font-size: 16px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 2px;
}
.ts-topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 阈值分组 */
.ts-group {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  padding: 14px 20px;
  margin-bottom: 12px;
  position: relative;
}
.ts-group::before {
  content: '';
  position: absolute; top: 0; left: 0; right: 0; height: 1px;
  background: linear-gradient(90deg, transparent, var(--cyan-ghost), transparent);
}

.ts-group-label {
  font-family: 'Orbitron', monospace;
  font-size: 12px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 2px;
  text-transform: uppercase;
  margin-bottom: 8px;
}

/* 分组内容缩进 */
.ts-group-inner {
  padding-left: 16px;
  border-left: 1px solid var(--bg-border2);
}

.ts-group-desc {
  font-size: 11px;
  color: var(--text-dim);
  margin: 0 0 12px 0;
  line-height: 1.4;
}

/* 子分组：时延 / 丢包率 */
.ts-subgroup {
  margin-bottom: 12px;
}
.ts-subgroup:last-child {
  margin-bottom: 0;
}

.ts-subgroup-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text);
  margin-bottom: 4px;
  padding-left: 8px;
  border-left: 2px solid var(--cyan-dim);
}

.ts-subgroup-desc {
  font-size: 10px;
  color: var(--text-dim);
  margin: 0 0 6px 0;
  line-height: 1.5;
}

/* 状态标签 */
.ts-tag {
  display: inline-block;
  padding: 0 5px;
  border-radius: 3px;
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: 600;
}
.ts-tag.ok { background: rgba(57,255,126,0.12); color: #39ff7e; }
.ts-tag.warn { background: rgba(255,171,0,0.12); color: #ffab00; }
.ts-tag.err { background: rgba(255,69,96,0.12); color: #ff4560; }

.ss-saved {
  font-size: 12px;
  color: var(--green);
}

/* 指标筛选区 */
.stats-topbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding-bottom: 8px;
  gap: 12px;
}
.stats-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 12px;
}
.stats-count {
  font-size: 11px;
  color: var(--text-dim);
  font-family: var(--font-mono);
}

/* 状态下拉颜色 */
.status-select.status-new     { color: var(--text-dim); border-color: var(--bg-border); }
.status-select.status-show    { color: var(--cyan);     border-color: var(--cyan-dim); }
.status-select.status-hide    { color: var(--warn);     border-color: rgba(255,171,0,0.3); }
.status-select.status-disable { color: var(--err);      border-color: rgba(255,69,96,0.3); }
.stats-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  padding-top: 2px;
}
.stats-filters {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.filter-label {
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 1px;
  text-transform: uppercase;
  min-width: 56px;
  font-family: var(--font-head);
}
.filter-btns {
  display: flex;
  gap: 4px;
}
.filter-btns .ctrl-btn.active {
  border-color: var(--cyan);
  color: var(--cyan);
  background: var(--cyan-ghost);
}

/* 机房配置表格 */
.idc-table {
  width: 100%;
}

.idc-thead {
  display: flex;
  padding: 0 0 10px 0;
  border-bottom: 1px solid var(--bg-border);
  margin-bottom: 4px;
}
.idc-th {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1px;
}
.idc-th.code { width: 280px; }
.idc-th.name { flex: 1; }
.idc-th.enable { width: 60px; text-align: center; }

.idc-tr {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--bg-border2);
}
.idc-tr:last-child {
  border-bottom: none;
}

.idc-td {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-dim);
}
.idc-td.code {
  width: 280px;
  color: var(--text);
  font-weight: 600;
}
.idc-td.name {
  flex: 1;
}
.idc-td.enable {
  width: 60px;
  display: flex;
  justify-content: center;
}

.idc-input {
  background: var(--bg-deep);
  border: 1px solid var(--bg-border);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  width: 100%;
  max-width: 220px;
  outline: none;
  transition: border-color 0.15s;
}
.idc-input:focus {
  border-color: var(--cyan-dim);
}

/* Toggle 开关 */
.idc-toggle {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
  cursor: pointer;
}
.idc-toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}
.idc-toggle-knob {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--bg-border);
  border-radius: 10px;
  transition: background 0.2s;
}
.idc-toggle-knob::before {
  content: '';
  position: absolute;
  top: 2px; left: 2px;
  width: 16px; height: 16px;
  background: var(--text-dim);
  border-radius: 50%;
  transition: transform 0.2s, background 0.2s;
}
.idc-toggle input:checked + .idc-toggle-knob {
  background: var(--cyan-ghost);
}
.idc-toggle input:checked + .idc-toggle-knob::before {
  transform: translateX(16px);
  background: var(--cyan);
}
</style>
