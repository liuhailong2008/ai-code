import { reactive } from 'vue'

// ── 节点数据 ──
export const nodes = reactive([
  { id:'CORE-RTR-01', type:'router',   status:'ok',   lat:4,  cpu:42, mem:56, ip:'10.0.0.1',  vendor:'Cisco',     model:'ASR 9001',  uptime:'187d' },
  { id:'CORE-RTR-02', type:'router',   status:'ok',   lat:6,  cpu:38, mem:52, ip:'10.0.0.2',  vendor:'Cisco',     model:'ASR 9001',  uptime:'187d' },
  { id:'CORE-SW-01',  type:'switch',   status:'warn', lat:12, cpu:78, mem:61, ip:'10.0.1.1',  vendor:'Huawei',    model:'CE 6870',   uptime:'94d'  },
  { id:'CORE-SW-02',  type:'switch',   status:'ok',   lat:8,  cpu:35, mem:44, ip:'10.0.1.2',  vendor:'Huawei',    model:'CE 6870',   uptime:'94d'  },
  { id:'FW-EDGE-01',  type:'firewall', status:'ok',   lat:3,  cpu:29, mem:67, ip:'10.0.2.1',  vendor:'Palo Alto', model:'PA-5250',   uptime:'210d' },
  { id:'EDGE-BJ-01',  type:'router',   status:'ok',   lat:14, cpu:22, mem:40, ip:'172.16.0.1', vendor:'Juniper',  model:'MX204',     uptime:'55d'  },
  { id:'EDGE-SH-01',  type:'router',   status:'err',  lat:0,  cpu:0,  mem:0,  ip:'172.16.1.1', vendor:'Juniper',  model:'MX204',     uptime:'失联' },
  { id:'EDGE-GZ-01',  type:'router',   status:'ok',   lat:22, cpu:18, mem:38, ip:'172.16.2.1', vendor:'Juniper',  model:'MX204',     uptime:'55d'  },
  { id:'EDGE-CD-01',  type:'router',   status:'warn', lat:55, cpu:51, mem:49, ip:'172.16.3.1', vendor:'Juniper',  model:'MX204',     uptime:'55d'  },
  { id:'EDGE-WH-01',  type:'router',   status:'ok',   lat:18, cpu:21, mem:36, ip:'172.16.4.1', vendor:'Juniper',  model:'MX204',     uptime:'55d'  },
])

export const events = reactive([
  { type:'err',  time:'14:22:07', host:'EDGE-SH-01',  msg:'接口 GigE0/0 链路中断，OSPF 邻居丢失' },
  { type:'warn', time:'14:21:44', host:'CORE-SW-01',  msg:'CPU 使用率 78%，超过阈值 70%' },
  { type:'warn', time:'14:20:31', host:'EDGE-CD-01',  msg:'VLAN 200 流量超基线 +140%' },
  { type:'ok',   time:'14:19:55', host:'CORE-RTR-01', msg:'BGP 会话重建，路由前缀 32,451' },
  { type:'ok',   time:'14:18:02', host:'FW-EDGE-01',  msg:'策略规则 #88 命中率清零重置' },
  { type:'warn', time:'14:16:14', host:'EDGE-BJ-01',  msg:'NTP 偏差 +38ms，超过告警阈值' },
  { type:'ok',   time:'14:15:00', host:'CORE-SW-02',  msg:'LACP 成员恢复，聚合带宽 20G' },
  { type:'err',  time:'14:13:30', host:'EDGE-SH-01',  msg:'ICMP 探针连续 5 次超时' },
])

export const alerts = [
  { level:'err',  title:'节点失联',  desc:'EDGE-SH-01 无法访问，ICMP/SNMP 全部超时',    time:'14:22:07', duration:'12m', host:'EDGE-SH-01',  rule:'连通性检测', impact:'3 条下游路由中断' },
  { level:'err',  title:'接口中断',  desc:'EDGE-SH-01 GigE0/0 物理链路断开',            time:'14:22:07', duration:'12m', host:'EDGE-SH-01',  rule:'接口监控',   impact:'上海节点完全失联' },
  { level:'warn', title:'CPU 高占',  desc:'CORE-SW-01 CPU 78%，阈值 70%',               time:'14:21:44', duration:'14m', host:'CORE-SW-01',  rule:'资源阈值',   impact:'转发性能下降' },
]

export const protos = [
  { name:'HTTP/3', pct:42, color:'#00f5c4' },
  { name:'QUIC',   pct:18, color:'#00c8f5' },
  { name:'TLS',    pct:15, color:'#39ff7e' },
  { name:'BGP',    pct:11, color:'#ffc233' },
  { name:'OSPF',   pct:8,  color:'#8e6fff' },
  { name:'其他',    pct:6,  color:'#3a6070' },
]

export const topSrc = [
  { ip:'10.12.44.8',    bw:'8.2 Gbps', pct:92 },
  { ip:'172.16.100.22', bw:'6.7 Gbps', pct:75 },
  { ip:'10.0.200.14',   bw:'5.9 Gbps', pct:66 },
  { ip:'192.168.1.1',   bw:'4.3 Gbps', pct:48 },
  { ip:'10.44.0.255',   bw:'3.8 Gbps', pct:43 },
  { ip:'172.31.8.4',    bw:'3.1 Gbps', pct:35 },
  { ip:'10.0.90.50',    bw:'2.6 Gbps', pct:29 },
]

// ── 拓扑数据 ──
export const topoNodes = [
  { id:'internet', x:0.5,  y:0.07, label:'WAN',          color:'#3a6070', r:18 },
  { id:'fw',       x:0.5,  y:0.22, label:'FW-EDGE-01',   color:'#00f5c4', r:15 },
  { id:'cr1',      x:0.28, y:0.40, label:'RTR-01',        color:'#00f5c4', r:14 },
  { id:'cr2',      x:0.72, y:0.40, label:'RTR-02',        color:'#00f5c4', r:14 },
  { id:'sw1',      x:0.14, y:0.62, label:'SW-01',         color:'#ffc233', r:13 },
  { id:'sw2',      x:0.5,  y:0.62, label:'SW-02',         color:'#00f5c4', r:13 },
  { id:'sw3',      x:0.86, y:0.62, label:'SW-03',         color:'#00f5c4', r:13 },
  { id:'e1',       x:0.06, y:0.86, label:'BJ',            color:'#39ff7e', r:10 },
  { id:'e2',       x:0.25, y:0.86, label:'SH',            color:'#ff4560', r:10 },
  { id:'e3',       x:0.44, y:0.86, label:'GZ',            color:'#39ff7e', r:10 },
  { id:'e4',       x:0.63, y:0.86, label:'CD',            color:'#ffc233', r:10 },
  { id:'e5',       x:0.82, y:0.86, label:'WH',            color:'#39ff7e', r:10 },
]

export const topoLinks = [
  ['internet','fw',true],['fw','cr1',true],['fw','cr2',true],
  ['cr1','sw1',true],['cr1','sw2',true],['cr2','sw2',true],['cr2','sw3',true],
  ['sw1','e1',true],['sw1','e2',false],['sw2','e3',true],['sw2','e4',false],['sw3','e5',true],
]

// ── 格式化时间 ──
export function formatTime() {
  const now = new Date()
  return now.toLocaleTimeString('zh-CN', { hour12: false })
}

export function formatDate() {
  const now = new Date()
  return now.toLocaleDateString('zh-CN', { month:'2-digit', day:'2-digit' }) + ' ' + formatTime()
}
