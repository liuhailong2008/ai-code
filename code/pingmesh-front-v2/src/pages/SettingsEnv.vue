<script setup>
import { ref, onMounted } from 'vue'
import { apiFetch } from '../utils/api.js'

const envNav = ['数据库配置', 'CAS配置', 'Prometheus配置']
const activeNav = ref(envNav[0])

// ── 数据库配置 ──
const dbForm = ref({ host: '', port: 3306, user: '', password: '', database: '' })
const dbSaving = ref(false)
const dbSaved = ref(false)
const dbTesting = ref(false)
const dbTestResult = ref('')
const dbTestError = ref('')
const dbTestAlert = ref(null)

async function loadDB() {
  try {
    const res = await apiFetch('/api/env-config/mysql')
    const data = await res.json()
    if (data.error) { console.error(data.error); return }
    dbForm.value = { ...data }
  } catch (e) {
    console.error('加载数据库配置失败:', e)
  }
}

async function saveDB() {
  dbSaving.value = true
  dbSaved.value = false
  try {
    const res = await apiFetch('/api/env-config/mysql', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(dbForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      dbSaved.value = true
      setTimeout(() => { dbSaved.value = false }, 3000)
    } else {
      alert('保存失败: ' + (data.error || '未知错误'))
    }
  } catch (e) {
    console.error('保存数据库配置失败:', e)
    alert('保存失败: ' + (e.message || e))
  } finally {
    dbSaving.value = false
  }
}

async function testDB() {
  dbTesting.value = true
  dbTestResult.value = ''
  dbTestError.value = ''
  dbTestAlert.value = null
  try {
    const res = await apiFetch('/api/env-config/mysql/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(dbForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      dbTestResult.value = data.message || '连接测试成功'
      if (data.alert) {
        dbTestAlert.value = data.alert
      }
    } else {
      dbTestError.value = data.error || '连接测试失败'
    }
  } catch (e) {
    dbTestError.value = '测试请求异常: ' + (e.message || e)
  } finally {
    dbTesting.value = false
  }
}

// ── CAS 配置 ──
const casForm = ref({
  enabled: false,
  server_url: '',
  login_url: '',
  validate_url: '',
  logout_url: '',
  service_url: '',
  frontend_callback: '',
  attribute_mapping: { username: '', display_name: '', email: '', phone: '', department: '' }
})
const casSaving = ref(false)
const casSaved = ref(false)
const casTesting = ref(false)
const casTestResult = ref('')
const casTestError = ref('')
const casTestResults = ref([])

async function loadCAS() {
  try {
    const res = await apiFetch('/api/env-config/cas')
    const data = await res.json()
    if (data.error) { console.error(data.error); return }
    casForm.value = {
      enabled: data.enabled || false,
      server_url: data.server_url || '',
      login_url: data.login_url || '',
      validate_url: data.validate_url || '',
      logout_url: data.logout_url || '',
      service_url: data.service_url || '',
      frontend_callback: data.frontend_callback || '',
      attribute_mapping: {
        username: data.attribute_mapping?.username || '',
        display_name: data.attribute_mapping?.display_name || '',
        email: data.attribute_mapping?.email || '',
        phone: data.attribute_mapping?.phone || '',
        department: data.attribute_mapping?.department || '',
      }
    }
  } catch (e) {
    console.error('加载CAS配置失败:', e)
  }
}

async function saveCAS() {
  casSaving.value = true
  casSaved.value = false
  try {
    const res = await apiFetch('/api/env-config/cas', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(casForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      casSaved.value = true
      setTimeout(() => { casSaved.value = false }, 3000)
    } else {
      alert('保存失败: ' + (data.error || '未知错误'))
    }
  } catch (e) {
    console.error('保存CAS配置失败:', e)
    alert('保存失败: ' + (e.message || e))
  } finally {
    casSaving.value = false
  }
}

async function testCAS() {
  casTesting.value = true
  casTestResult.value = ''
  casTestError.value = ''
  casTestResults.value = []
  try {
    const res = await apiFetch('/api/env-config/cas/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(casForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok' || data.status === 'partial') {
      casTestResult.value = data.message || '测试完成'
      casTestResults.value = data.results || []
    } else {
      casTestError.value = data.error || '连接测试失败'
    }
  } catch (e) {
    casTestError.value = '测试请求异常: ' + (e.message || e)
  } finally {
    casTesting.value = false
  }
}

// ── Prometheus 配置 ──
const promForm = ref({ base_url: '' })
const promSaving = ref(false)
const promSaved = ref(false)
const promTesting = ref(false)
const promTestResult = ref('')
const promTestError = ref('')

async function loadProm() {
  try {
    const res = await apiFetch('/api/env-config/prometheus')
    const data = await res.json()
    if (data.error) { console.error(data.error); return }
    promForm.value = { base_url: data.base_url || '' }
  } catch (e) {
    console.error('加载Prometheus配置失败:', e)
  }
}

async function saveProm() {
  promSaving.value = true
  promSaved.value = false
  try {
    const res = await apiFetch('/api/env-config/prometheus', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(promForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      promSaved.value = true
      setTimeout(() => { promSaved.value = false }, 3000)
    } else {
      alert('保存失败: ' + (data.error || '未知错误'))
    }
  } catch (e) {
    console.error('保存Prometheus配置失败:', e)
    alert('保存失败: ' + (e.message || e))
  } finally {
    promSaving.value = false
  }
}

async function testProm() {
  promTesting.value = true
  promTestResult.value = ''
  promTestError.value = ''
  try {
    const res = await apiFetch('/api/env-config/prometheus/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(promForm.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      promTestResult.value = data.message || '连接测试成功'
    } else {
      promTestError.value = data.error || '连接测试失败'
    }
  } catch (e) {
    promTestError.value = '测试请求异常: ' + (e.message || e)
  } finally {
    promTesting.value = false
  }
}

onMounted(() => {
  loadDB()
  loadCAS()
  loadProm()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">环境管理</h2>
    </div>
    <div class="settings-layout">
      <div class="settings-nav">
        <div
          v-for="item in envNav" :key="item"
          class="sn-item" :class="{active:activeNav===item}"
          @click="activeNav = item"
        >{{ item }}</div>
      </div>
      <div class="settings-content">

        <!-- 数据库配置 Tab -->
        <div v-if="activeNav==='数据库配置'" class="settings-section">
          <div class="ts-topbar">
            <span class="ts-topbar-title">MySQL 数据库配置</span>
            <div class="ts-topbar-actions">
              <button class="ctrl-btn" :disabled="dbTesting" @click="testDB">
                {{ dbTesting ? '测试中...' : '测试连接' }}
              </button>
              <button class="ctrl-btn accent-btn" :disabled="dbSaving" @click="saveDB">
                {{ dbSaving ? '保存中...' : '保存配置' }}
              </button>
              <span v-if="dbSaved" class="ss-saved">保存成功</span>
            </div>
          </div>

          <!-- 测试结果 -->
          <div v-if="dbTestResult" class="test-result ok">{{ dbTestResult }}</div>
          <div v-if="dbTestError" class="test-result err">{{ dbTestError }}</div>
          <div v-if="dbTestAlert" class="test-result ok">
            <div class="alert-preview">
              <div class="alert-preview-title">alert_record 最新告警数据：</div>
              <div class="alert-preview-grid">
                <div><span>ID</span><span>{{ dbTestAlert.id }}</span></div>
                <div><span>故障机房</span><span>{{ dbTestAlert.alert_idc }}</span></div>
                <div><span>故障Spine</span><span>{{ dbTestAlert.alert_spine || '-' }}</span></div>
                <div><span>故障Leaf</span><span>{{ dbTestAlert.alert_leaf || '-' }}</span></div>
                <div><span>故障机柜</span><span>{{ dbTestAlert.alert_rack || '-' }}</span></div>
                <div><span>告警类型</span><span>{{ dbTestAlert.alert_type || '-' }}</span></div>
                <div><span>告警状态</span><span>{{ dbTestAlert.alert_status || '-' }}</span></div>
                <div><span>创建人</span><span>{{ dbTestAlert.create_by || '-' }}</span></div>
                <div><span>创建时间</span><span>{{ dbTestAlert.create_time }}</span></div>
                <div><span>更新人</span><span>{{ dbTestAlert.update_by || '-' }}</span></div>
                <div><span>更新时间</span><span>{{ dbTestAlert.update_time }}</span></div>
              </div>
            </div>
          </div>

          <div class="env-form">
            <div class="env-field">
              <label>主机地址</label>
              <input class="ctrl-input" v-model="dbForm.host" placeholder="127.0.0.1" />
            </div>
            <div class="env-field">
              <label>端口</label>
              <input class="ctrl-input" v-model.number="dbForm.port" type="number" placeholder="3306" />
            </div>
            <div class="env-field">
              <label>用户名</label>
              <input class="ctrl-input" v-model="dbForm.user" placeholder="root" />
            </div>
            <div class="env-field">
              <label>密码</label>
              <input class="ctrl-input" v-model="dbForm.password" type="password" placeholder="输入密码" />
              <span class="field-hint">密码不回显，已保存的密码显示为 ******</span>
            </div>
            <div class="env-field">
              <label>数据库名</label>
              <input class="ctrl-input" v-model="dbForm.database" placeholder="pingmesh" />
            </div>
          </div>
        </div>

        <!-- CAS 配置 Tab -->
        <div v-if="activeNav==='CAS配置'" class="settings-section">
          <div class="ts-topbar">
            <span class="ts-topbar-title">CAS 单点登录配置</span>
            <div class="ts-topbar-actions">
              <button class="ctrl-btn" :disabled="casTesting" @click="testCAS">
                {{ casTesting ? '测试中...' : '测试连接' }}
              </button>
              <button class="ctrl-btn accent-btn" :disabled="casSaving" @click="saveCAS">
                {{ casSaving ? '保存中...' : '保存配置' }}
              </button>
              <span v-if="casSaved" class="ss-saved">保存成功</span>
            </div>
          </div>

          <div v-if="casTestResult" class="test-result" :class="casTestResults.some(r => r.status === 'fail') ? 'err' : 'ok'">{{ casTestResult }}</div>
          <div v-if="casTestError" class="test-result err">{{ casTestError }}</div>
          <div v-if="casTestResults.length" class="test-result ok">
            <div class="alert-preview">
              <div class="alert-preview-title">各链接测试结果：</div>
              <div v-for="r in casTestResults" :key="r.name" class="cas-result-row">
                <span class="cas-result-status" :class="r.status">{{ r.status === 'ok' ? '✓' : '✗' }}</span>
                <div class="cas-result-info">
                  <span class="cas-result-name">{{ r.name }}</span>
                  <span class="cas-result-url">{{ r.url || '(未配置)' }}</span>
                  <span class="cas-result-error" v-if="r.status === 'fail'">{{ r.error }}</span>
                  <span class="cas-result-error" v-else style="color:#39ff7e;">{{ r.error }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="env-form">
            <div class="env-field">
              <label>
                启用 CAS
                <span class="field-hint">（开启后将禁用本地登录）</span>
              </label>
              <label class="idc-toggle" style="display:inline-flex;">
                <input type="checkbox" v-model="casForm.enabled" />
                <span class="idc-toggle-knob"></span>
              </label>
            </div>
            <div class="env-field">
              <label>CAS 服务器地址</label>
              <input class="ctrl-input" v-model="casForm.server_url" placeholder="https://cas.example.com" />
            </div>
            <div class="env-field">
              <label>登录页 URL</label>
              <input class="ctrl-input" v-model="casForm.login_url" placeholder="https://cas.example.com/login" />
            </div>
            <div class="env-field">
              <label>验证票据 URL</label>
              <input class="ctrl-input" v-model="casForm.validate_url" placeholder="https://cas.example.com/serviceValidate" />
            </div>
            <div class="env-field">
              <label>登出 URL</label>
              <input class="ctrl-input" v-model="casForm.logout_url" placeholder="https://cas.example.com/logout" />
            </div>
            <div class="env-field">
              <label>Service URL（回调地址）</label>
              <input class="ctrl-input" v-model="casForm.service_url" placeholder="http://your-domain/api/auth/cas/callback" />
            </div>
            <div class="env-field">
              <label>前端回调地址</label>
              <input class="ctrl-input" v-model="casForm.frontend_callback" placeholder="http://your-domain/#/cas/callback" />
            </div>
          </div>

          <div class="ts-subsection">
            <div class="ts-subsection-title">属性映射</div>
            <div class="env-form">
              <div class="env-field">
                <label>用户名字段</label>
                <input class="ctrl-input" v-model="casForm.attribute_mapping.username" placeholder="uid" />
              </div>
              <div class="env-field">
                <label>显示名称字段</label>
                <input class="ctrl-input" v-model="casForm.attribute_mapping.display_name" placeholder="displayName" />
              </div>
              <div class="env-field">
                <label>邮箱字段</label>
                <input class="ctrl-input" v-model="casForm.attribute_mapping.email" placeholder="mail" />
              </div>
              <div class="env-field">
                <label>电话字段</label>
                <input class="ctrl-input" v-model="casForm.attribute_mapping.phone" placeholder="mobile" />
              </div>
              <div class="env-field">
                <label>部门字段</label>
                <input class="ctrl-input" v-model="casForm.attribute_mapping.department" placeholder="department" />
              </div>
            </div>
          </div>
        </div>

        <!-- Prometheus 配置 Tab -->
        <div v-if="activeNav==='Prometheus配置'" class="settings-section">
          <div class="ts-topbar">
            <span class="ts-topbar-title">Prometheus 配置</span>
            <div class="ts-topbar-actions">
              <button class="ctrl-btn" :disabled="promTesting" @click="testProm">
                {{ promTesting ? '测试中...' : '测试连接' }}
              </button>
              <button class="ctrl-btn accent-btn" :disabled="promSaving" @click="saveProm">
                {{ promSaving ? '保存中...' : '保存配置' }}
              </button>
              <span v-if="promSaved" class="ss-saved">保存成功</span>
            </div>
          </div>

          <div v-if="promTestResult" class="test-result ok">{{ promTestResult }}</div>
          <div v-if="promTestError" class="test-result err">{{ promTestError }}</div>

          <div class="env-form">
            <div class="env-field">
              <label>Prometheus Base URL</label>
              <input class="ctrl-input" v-model="promForm.base_url" placeholder="http://127.0.0.1:9090" style="width:400px;" />
              <span class="field-hint">Prometheus API 地址，用于查询实时监控指标</span>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.env-form {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-top: 12px;
}

.env-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.env-field label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1px;
  font-family: var(--font-head);
  display: flex;
  align-items: center;
  gap: 12px;
}

.env-field .ctrl-input {
  max-width: 360px;
}

.field-hint {
  font-size: 10px;
  color: var(--text-dim);
  font-weight: 400;
  text-transform: none;
  letter-spacing: 0;
}

.test-result {
  margin-top: 8px;
  padding: 10px 14px;
  border-radius: var(--radius);
  font-size: 12px;
  font-family: var(--font-mono);
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}
.test-result.ok {
  background: rgba(57,255,126,0.08);
  border: 1px solid rgba(57,255,126,0.2);
  color: #39ff7e;
}
.test-result.err {
  background: rgba(255,69,96,0.08);
  border: 1px solid rgba(255,69,96,0.2);
  color: #ff4560;
}

.ts-subsection {
  margin-top: 18px;
  padding-top: 14px;
  border-top: 1px solid var(--bg-border2);
}

.ts-subsection-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--cyan);
  font-family: var(--font-head);
  letter-spacing: 1px;
  margin-bottom: 10px;
}

.alert-preview {
  white-space: normal;
  word-break: normal;
}

.alert-preview-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--cyan);
  margin-bottom: 8px;
}

.alert-preview-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px 16px;
}

.alert-preview-grid > div {
  display: flex;
  justify-content: space-between;
  padding: 3px 0;
  border-bottom: 1px solid rgba(57,255,126,0.08);
}

.alert-preview-grid > div > span:first-child {
  color: var(--text-dim);
  font-size: 11px;
}

.alert-preview-grid > div > span:last-child {
  color: var(--text);
  font-size: 11px;
  font-weight: 500;
}

.cas-result-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 6px 0;
  border-bottom: 1px solid rgba(57,255,126,0.06);
}

.cas-result-row:last-child {
  border-bottom: none;
}

.cas-result-status {
  font-size: 14px;
  font-weight: 700;
  min-width: 20px;
  text-align: center;
  padding-top: 1px;
}

.cas-result-status.ok {
  color: #39ff7e;
}

.cas-result-status.fail {
  color: #ff4560;
}

.cas-result-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.cas-result-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
}

.cas-result-url {
  font-size: 10px;
  color: var(--text-dim);
  font-family: var(--font-mono);
  word-break: break-all;
}

.cas-result-error {
  font-size: 10px;
  color: #ff4560;
  font-family: var(--font-mono);
}
</style>
