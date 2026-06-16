<script setup>
import { ref, computed } from 'vue'
import { nodes } from '@/composables/useData'

const search = ref('')
const typeFilter = ref('全部类型')
const types = ['全部类型', '路由器', '交换机', '防火墙']
const typeMap = { router: '路由器', switch: '交换机', firewall: '防火墙' }
const statusLabels = { ok: '正常', warn: '警告', err: '故障' }

const filteredNodes = computed(() => {
  let data = nodes
  if (search.value) {
    const q = search.value.toLowerCase()
    data = data.filter(n => n.id.toLowerCase().includes(q) || n.ip.includes(q))
  }
  if (typeFilter.value !== '全部类型') {
    const map = { '路由器': 'router', '交换机': 'switch', '防火墙': 'firewall' }
    data = data.filter(n => n.type === map[typeFilter.value])
  }
  return data
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">设备管理</h2>
      <div class="page-controls">
        <input class="ctrl-input" placeholder="搜索设备 IP / 名称..." v-model="search" />
        <select class="ctrl-select" v-model="typeFilter">
          <option v-for="t in types" :key="t">{{ t }}</option>
        </select>
        <button class="ctrl-btn accent-btn">+ 添加设备</button>
      </div>
    </div>
    <div class="devices-table-wrap">
      <table class="devices-table">
        <thead><tr>
          <th>设备名称</th><th>类型</th><th>厂商 / 型号</th><th>IP 地址</th>
          <th>状态</th><th>CPU</th><th>内存</th><th>运行时间</th><th>操作</th>
        </tr></thead>
        <tbody>
          <tr v-for="d in filteredNodes" :key="d.id">
            <td><div class="dev-name">{{ d.id }}</div></td>
            <td><span class="dev-vendor">{{ typeMap[d.type] || d.type }}</span></td>
            <td><span class="dev-vendor">{{ d.vendor }} {{ d.model }}</span></td>
            <td><span class="dev-ip">{{ d.ip }}</span></td>
            <td><span class="status-badge" :class="d.status">{{ statusLabels[d.status] }}</span></td>
            <td>
              <div
                :style="{fontSize:'11px',color:d.cpu>70?'#ff4560':d.cpu>50?'#ffc233':'#39ff7e'}"
              >{{ d.cpu || 0 }}%</div>
              <div class="usage-bar">
                <div class="usage-fill" :style="{width:(d.cpu||0)+'%',background:d.cpu>70?'#ff4560':d.cpu>50?'#ffc233':'#39ff7e'}"></div>
              </div>
            </td>
            <td>
              <div
                :style="{fontSize:'11px',color:d.mem>80?'#ff4560':d.mem>60?'#ffc233':'#00f5c4'}"
              >{{ d.mem || 0 }}%</div>
              <div class="usage-bar">
                <div class="usage-fill" :style="{width:(d.mem||0)+'%',background:d.mem>80?'#ff4560':d.mem>60?'#ffc233':'#00f5c4'}"></div>
              </div>
            </td>
            <td style="color:var(--text-dim);font-size:11px">{{ d.uptime }}</td>
            <td>
              <div class="dev-actions">
                <button class="dev-btn">终端</button>
                <button class="dev-btn">详情</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
