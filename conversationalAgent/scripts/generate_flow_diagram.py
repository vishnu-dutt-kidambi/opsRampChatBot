#!/usr/bin/env python3
"""
Generate HPE Autopilot flow diagram — accurate to the GreenLake portal demo scenario.

Layout:
  ┌─────────────────────────────────────────────────────────┐
  │  Title + subtitle                                       │
  │  ┌─ Web UI ─┐  ┌─ CLI ─┐  ┌─ MCP Server ─┐            │
  │  └──────────┘  └───────┘  └──────────────┘             │
  │         ↓          ↓            ↓                       │
  │  ┌── USER QUESTION ─────────────────────────┐           │
  │  │ "Why is the GreenLake portal slow?"      │           │
  │  └──────────────────────────────────────────┘           │
  │                    ↓                                    │
  │  ┌── ReAct Engine (Ollama + Llama 3.1) ────┐           │
  │  └─────────────────────────────────────────┘            │
  │                    ↓                                    │
  │  ┌╌╌╌ AUTONOMOUS REASONING LOOP ╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┐  │
  │  ╎  [Round 1-6 left]  ←---→  [Data sources right]   ╎  │
  │  └╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌┘  │
  │                    ↓                                    │
  │  ┌── RESOLUTION ───────────────────────────┐            │
  │  └─────────────────────────────────────────┘            │
  │  ┌── Legend ───────────────────────────────┐            │
  └─────────────────────────────────────────────────────────┘
"""

import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from matplotlib.patches import FancyBboxPatch

# ─── Canvas ────────────────────────────────────────────────
W, H = 20, 32
fig, ax = plt.subplots(figsize=(W * 0.9, H * 0.9))
ax.set_xlim(0, W)
ax.set_ylim(0, H)
ax.set_aspect('equal')
ax.axis('off')
fig.patch.set_facecolor('#ffffff')

# ─── Grid constants ────────────────────────────────────────
CX        = 10.0          # global center X
LEFT_CX   = 6.5           # center of left column (rounds)
RIGHT_CX  = 15.5          # center of right column (sources)
BOX_W     = 7.6           # round box width
BOX_H     = 1.5           # round box height
SRC_W     = 4.0           # source box width
SRC_H     = 1.3           # source box height
GAP       = 0.5           # vertical gap between boxes
ARROW_GAP = 0.15          # gap between box edge and arrow tip

# ─── Helpers ───────────────────────────────────────────────
def box(x, y, w, h, fc, ec, lw=2.0, ls='-'):
    """Draw a rounded box. (x, y) is bottom-left."""
    ax.add_patch(FancyBboxPatch(
        (x, y), w, h, boxstyle="round,pad=0.25",
        facecolor=fc, edgecolor=ec, linewidth=lw, linestyle=ls))

def txt(x, y, s, fs=11, fw='normal', fc='#1e293b', ha='center', **kw):
    ax.text(x, y, s, ha=ha, va='center',
            fontsize=fs, fontweight=fw, color=fc,
            fontfamily='sans-serif', **kw)

def arrow(x1, y1, x2, y2, c='#64748b'):
    ax.annotate('', xy=(x2, y2), xytext=(x1, y1),
                arrowprops=dict(arrowstyle='->', color=c, lw=2.2,
                                shrinkA=0, shrinkB=0))

def dash(x1, y1, x2, y2, c='#94a3b8'):
    ax.plot([x1, x2], [y1, y2], linestyle=(0, (6, 4)),
            color=c, lw=1.6, alpha=0.6, solid_capstyle='round')

# ─── Y positions (top-down, calculated precisely) ──────────
Y_TITLE    = 31.0
Y_SUBTITLE = 30.4

# Interface boxes row
Y_IFACE    = 29.2
IFACE_H    = 0.8

# User question
Y_USER_TOP = 28.0
USER_H     = 1.3
Y_USER_BOT = Y_USER_TOP - USER_H

# Engine
Y_ENGINE_TOP = Y_USER_BOT - GAP - 0.1
ENGINE_H     = 1.1
Y_ENGINE_BOT = Y_ENGINE_TOP - ENGINE_H

# Loop
Y_LOOP_TOP   = Y_ENGINE_BOT - GAP - 0.1
ROUND_TOTAL  = 6 * BOX_H + 5 * GAP                        # 11.5
Y_LOOP_PAD   = 0.9
Y_LOOP_BOT   = Y_LOOP_TOP - ROUND_TOTAL - 2 * Y_LOOP_PAD
LOOP_H       = Y_LOOP_TOP - Y_LOOP_BOT

# Round Y positions (bottom-left y of each round box)
Y_R = []
for i in range(6):
    Y_R.append(Y_LOOP_TOP - Y_LOOP_PAD - i * (BOX_H + GAP) - BOX_H)

# Result
Y_RESULT_TOP = Y_LOOP_BOT - GAP - 0.3
RESULT_H     = 1.5
Y_RESULT_BOT = Y_RESULT_TOP - RESULT_H

Y_LEGEND     = Y_RESULT_BOT - 0.8

# ═══════════════════════════════════════════════════════════
# DRAW
# ═══════════════════════════════════════════════════════════

# ── Title ──────────────────────────────────────────────────
txt(CX, Y_TITLE, 'HPE Autopilot', fs=26, fw='bold', fc='#0f172a')
txt(CX, Y_SUBTITLE, 'Autonomous IT Incident Resolution  —  ReAct AI  |  11 Tools  |  Full-Stack Diagnosis',
    fs=12, fc='#64748b', style='italic')

# ── 0. Interface boxes ────────────────────────────────────
ifaces = [
    ("Web UI",     5.5,  '#dbeafe', '#3b82f6'),
    ("CLI",        9.2,  '#dbeafe', '#3b82f6'),
    ("MCP Server", 12.5, '#dbeafe', '#3b82f6'),
]
for label, ix, ifc, iec in ifaces:
    iw = 2.6
    box(ix - iw / 2, Y_IFACE - IFACE_H / 2, iw, IFACE_H, ifc, iec, lw=1.8)
    txt(ix, Y_IFACE, label, fs=10, fw='bold', fc='#1e40af')

# Arrows from interfaces down to user box
for _, ix, _, iec in ifaces:
    arrow(ix, Y_IFACE - IFACE_H / 2 - ARROW_GAP,
          CX, Y_USER_TOP + ARROW_GAP, '#3b82f6')

# ── 1. User box ───────────────────────────────────────────
uw = 10.0
ux = CX - uw / 2
box(ux, Y_USER_BOT, uw, USER_H, '#4f46e5', '#3730a3', lw=2.5)
txt(CX, Y_USER_BOT + USER_H * 0.65, 'USER QUESTION',
    fs=14, fw='bold', fc='white')
txt(CX, Y_USER_BOT + USER_H * 0.28, '"Why is the GreenLake portal slow?"',
    fs=11.5, fc='#c7d2fe', style='italic')

# Arrow User -> Engine
arrow(CX, Y_USER_BOT - ARROW_GAP,
      CX, Y_ENGINE_TOP + ARROW_GAP, '#4f46e5')

# ── 2. Engine box ─────────────────────────────────────────
ew = 12.0
ex = CX - ew / 2
box(ex, Y_ENGINE_BOT, ew, ENGINE_H, '#7c3aed', '#5b21b6', lw=2.5)
txt(CX, Y_ENGINE_BOT + ENGINE_H * 0.62,
    'HPE Autopilot  —  ReAct AI Engine',
    fs=13, fw='bold', fc='white')
txt(CX, Y_ENGINE_BOT + ENGINE_H * 0.28,
    'Ollama  +  Llama 3.1  |  Think → Act → Observe → Repeat',
    fs=10, fc='#ddd6fe')

# Arrow Engine -> Loop
arrow(CX, Y_ENGINE_BOT - ARROW_GAP,
      CX, Y_LOOP_TOP + ARROW_GAP, '#7c3aed')

# ── 3. Loop boundary ──────────────────────────────────────
loop_w = 17.4
loop_x = CX - loop_w / 2
box(loop_x, Y_LOOP_BOT, loop_w, LOOP_H, '#f8fafc', '#cbd5e1', lw=2, ls='--')
txt(CX, Y_LOOP_TOP - 0.35, 'AUTONOMOUS REASONING LOOP  (max 12 rounds)',
    fs=14, fw='bold', fc='#475569')

# ── 4. Round boxes (left column) ──────────────────────────
# Each round uses REAL mock data from the GreenLake portal scenario:
#   - Alert: ALR-20260219-009 HTTP 6.2s on k8s-node-04
#   - Server: k8s-node-04 (HPE ProLiant DL380)
#   - Network: sw-dc-east-04 ge-0/0/5 — 8.3% loss, duplex mismatch
#   - Blast radius: 3 apps, 5,000 users
rounds = [
    ("ROUND 1   search_alerts",
     "HTTP 6.2s latency + container restarts on k8s-node-04",
     '#6366f1', '#eef2ff'),
    ("ROUND 2   investigate_resource",
     "k8s-node-04: CPU 34%, Mem 72%, Disk 45% — all normal",
     '#6366f1', '#eef2ff'),
    ("ROUND 3   correlate_network",
     "sw-dc-east-04 ge-0/0/5: 8.3% loss, 120ms, duplex mismatch",
     '#0ea5e9', '#ecfeff'),
    ("ROUND 4   blast_radius",
     "3 apps + 5,000 users impacted (GreenLake, Aruba, DSCC)",
     '#f59e0b', '#fffbeb'),
    ("ROUND 5   search_knowledge_base",
     "RAG: Juniper switch port troubleshooting runbook found",
     '#8b5cf6', '#f5f3ff'),
    ("ROUND 6   get_remediation_plan",
     "Fix: interface reset + duplex auto → approval gate",
     '#10b981', '#ecfdf5'),
]

rx = LEFT_CX - BOX_W / 2

for i, (title, detail, bc, fc) in enumerate(rounds):
    by = Y_R[i]
    box(rx, by, BOX_W, BOX_H, fc, bc, lw=2.5)
    txt(rx + BOX_W / 2, by + BOX_H * 0.68, title,
        fs=11, fw='bold', fc='#1e293b')
    txt(rx + BOX_W / 2, by + BOX_H * 0.30, detail,
        fs=9.5, fc='#64748b', style='italic')
    # Arrow to next round
    if i < 5:
        arrow(LEFT_CX, by - ARROW_GAP,
              LEFT_CX, Y_R[i + 1] + BOX_H + ARROW_GAP, bc)

# ── 5. Data source boxes (right column) ───────────────────
txt(RIGHT_CX, Y_LOOP_TOP - 0.35, 'DATA SOURCES',
    fs=13, fw='bold', fc='#475569')

src_sx = RIGHT_CX - SRC_W / 2

sources = [
    ("OpsRamp API v2\nAlerts | Resources\nIncidents | Metrics",
     '#eef2ff', '#6366f1', 0),    # aligned to rounds 1-2
    ("Juniper Mist API\nSwitch Ports | ARP\nMAC | Telemetry",
     '#ecfeff', '#0ea5e9', 2),    # aligned to round 3
    ("Dependency Graph\nNetwork ↔ Server\n↔ Application",
     '#fffbeb', '#f59e0b', 3),    # aligned to round 4
    ("Runbook PDFs\nRAG Vector Store\n(nomic-embed-text)",
     '#f5f3ff', '#8b5cf6', 4),    # aligned to round 5
]

for label, fc, bc, ri in sources:
    round_cy = Y_R[ri] + BOX_H / 2
    sy = round_cy - SRC_H / 2
    box(src_sx, sy, SRC_W, SRC_H, fc, bc, lw=2)
    txt(RIGHT_CX, round_cy, label,
        fs=9.5, fw='bold', fc='#334155', linespacing=1.3)
    # Dashed line from round box to source box
    dash(rx + BOX_W, round_cy, src_sx, round_cy, bc)

# Also connect round 1 (alerts) source to round 2 (investigate) — same OpsRamp API
round1_cy = Y_R[1] + BOX_H / 2
dash(rx + BOX_W, round1_cy, src_sx, Y_R[0] + BOX_H / 2, '#6366f1')

# ── 6. Arrow last round -> result ─────────────────────────
arrow(LEFT_CX, Y_R[5] - ARROW_GAP,
      CX, Y_RESULT_TOP + ARROW_GAP, '#10b981')

# ── 7. Resolution box ─────────────────────────────────────
res_w = 16.0
res_x = CX - res_w / 2
box(res_x, Y_RESULT_BOT, res_w, RESULT_H, '#059669', '#047857', lw=2.5)
txt(CX, Y_RESULT_BOT + RESULT_H * 0.72, 'RESOLUTION  —  COMPREHENSIVE INCIDENT REPORT',
    fs=14, fw='bold', fc='white')
txt(CX, Y_RESULT_BOT + RESULT_H * 0.42,
    'Root Cause: Juniper sw-dc-east-04 ge-0/0/5 — 8.3% packet loss, duplex mismatch',
    fs=9.5, fc='#d1fae5')
txt(CX, Y_RESULT_BOT + RESULT_H * 0.18,
    'Impact: 3 apps, 5,000 users   |   Fix: interface reset + duplex auto   |   Status: awaiting approval',
    fs=9.5, fc='#d1fae5')

# ── 8. Legend ──────────────────────────────────────────────
# Solid arrow
lx1 = 3.0
ax.annotate('', xy=(lx1 + 1.8, Y_LEGEND), xytext=(lx1, Y_LEGEND),
            arrowprops=dict(arrowstyle='->', color='#64748b', lw=2))
txt(lx1 + 2.2, Y_LEGEND, 'Reasoning flow', fs=10, fc='#64748b', ha='left')

# Dashed line
lx2 = 8.5
ax.plot([lx2, lx2 + 1.5], [Y_LEGEND, Y_LEGEND],
        linestyle=(0, (6, 4)), color='#94a3b8', lw=1.6)
txt(lx2 + 1.9, Y_LEGEND, 'Data lookup', fs=10, fc='#94a3b8', ha='left')

# Interface badge
lx3 = 13.5
box(lx3, Y_LEGEND - 0.25, 1.2, 0.5, '#dbeafe', '#3b82f6', lw=1.2)
txt(lx3 + 1.6 + 0.3, Y_LEGEND, 'Interface', fs=10, fc='#3b82f6', ha='left')

# ── Save ───────────────────────────────────────────────────
plt.subplots_adjust(left=0.02, right=0.98, top=0.98, bottom=0.02)
out = '/Users/a60168034/go/src/ai/opsRampChatBot/conversationalAgent/screenshots/hpe_autopilot_flow.png'
plt.savefig(out, dpi=200, bbox_inches='tight', facecolor='white', edgecolor='none')
print(f"Saved: {out}")

# Also save to the old name for backward compat
out2 = '/Users/a60168034/go/src/ai/opsRampChatBot/conversationalAgent/screenshots/opsramp_autopilot_flow.png'
plt.savefig(out2, dpi=200, bbox_inches='tight', facecolor='white', edgecolor='none')
print(f"Saved: {out2}")
