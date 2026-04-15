import{b as l,c as i,d as t,F as b,r as v,t as a,h as g,n as f,f as h,e as x,g as _,i as o}from"./index-pKmMBxQM.js";import{_ as k}from"./_plugin-vue_export-helper-DlAUqK2U.js";const C={class:"page-header animate-in"},w={class:"page-actions"},L={class:"card animate-in delay-1"},P={class:"card-body",style:{padding:"0"}},z={class:"data-table"},H={class:"node-name"},S=["onClick"],$={class:"card animate-in delay-2",style:{"margin-top":"20px"}},A={class:"card-body"},B={__name:"AlertConfig",setup(D){const r=o(!1),c=o([{id:1,name:"链路延迟告警",type:"latency",condition:"> 250ms",duration:"5分钟",severity:"critical",enabled:!0},{id:2,name:"链路抖动告警",type:"jitter",condition:"> 50ms",duration:"10分钟",severity:"warning",enabled:!0},{id:3,name:"丢包率告警",type:"loss",condition:"> 1%",duration:"5分钟",severity:"critical",enabled:!0},{id:4,name:"节点离线告警",type:"offline",condition:"心跳超时",duration:"3分钟",severity:"critical",enabled:!0},{id:5,name:"P99延迟告警",type:"p99",condition:"> 200ms",duration:"15分钟",severity:"warning",enabled:!1}]),d=o(`groups:
- name: pingmesh_alerts
  rules:
  - alert: HighLatency
    expr: pingmesh_latency > 250
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High latency detected"
      description: "Link latency is {{ $value }}ms"

  - alert: HighPacketLoss
    expr: pingmesh_loss_rate > 1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High packet loss"
      description: "Packet loss rate is {{ $value }}%"`);function u(s){return s==="critical"?"#ef4444":s==="warning"?"#f59e0b":"#3b82f6"}function p(s){return s==="critical"?"严重":s==="warning"?"警告":"通知"}function m(s){s.enabled=!s.enabled}function y(){alert("规则保存成功！")}return(s,e)=>(l(),i("div",null,[t("div",C,[e[2]||(e[2]=t("div",null,[t("div",{class:"page-title"},"告警配置"),t("div",{class:"page-subtitle"},"配置告警规则和阈值")],-1)),t("div",w,[t("button",{class:"btn btn-primary",onClick:e[0]||(e[0]=n=>r.value=!0)},"+ 添加规则")])]),t("div",L,[e[6]||(e[6]=t("div",{class:"card-header"},[t("span",{class:"card-title"},"告警规则列表")],-1)),t("div",P,[t("table",z,[e[5]||(e[5]=t("thead",null,[t("tr",null,[t("th",null,"规则名称"),t("th",null,"告警类型"),t("th",null,"阈值条件"),t("th",null,"持续时间"),t("th",null,"严重性"),t("th",null,"状态"),t("th",null,"操作")])],-1)),t("tbody",null,[(l(!0),i(b,null,v(c.value,n=>(l(),i("tr",{key:n.id},[t("td",H,a(n.name),1),t("td",null,a(n.type),1),t("td",null,a(n.condition),1),t("td",null,a(n.duration),1),t("td",null,[t("span",{style:g({color:u(n.severity),fontWeight:600})},a(p(n.severity)),5)]),t("td",null,[t("span",{class:f(["status-dot",n.enabled?"online":"offline"])},[e[3]||(e[3]=t("span",{class:"dot"},null,-1)),h(" "+a(n.enabled?"启用":"禁用"),1)],2)]),t("td",null,[t("button",{class:"btn btn-outline",style:{padding:"4px 8px","font-size":"11px","margin-right":"4px"},onClick:R=>m(n)},a(n.enabled?"禁用":"启用"),9,S),e[4]||(e[4]=t("button",{class:"btn btn-outline",style:{padding:"4px 8px","font-size":"11px"}},"编辑",-1))])]))),128))])])])]),t("div",$,[e[7]||(e[7]=t("div",{class:"card-header"},[t("span",{class:"card-title"},"Prometheus 告警规则配置")],-1)),t("div",A,[x(t("textarea",{class:"rule-editor","onUpdate:modelValue":e[1]||(e[1]=n=>d.value=n),placeholder:"输入 Prometheus 告警规则..."},null,512),[[_,d.value]])]),t("div",{class:"card-footer",style:{padding:"12px 22px","border-top":"1px solid var(--border-color)"}},[t("button",{class:"btn btn-primary",onClick:y},"保存配置")])])]))}},N=k(B,[["__scopeId","data-v-d4557c6a"]]);export{N as default};
