#!/usr/bin/env python3
"""Generate OpsRamp Autopilot flow diagram — precise grid-aligned layout."""

import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
from matplotlib.patches import FancyBboxPatch

# ─── Canvas ────────────────────────────────────────────────
W, H = 20, 28
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
Y_TITLE    = 27.0
Y_SUBTITLE = 26.4
Y_USER_TOP = 25.6
USER_H     = 1.2
Y_USER_BOT = Y_USER_TOP - USER_H                        # 24.4

Y_ENGINE_TOP = Y_USER_BOT - GAP - 0.1                   # 23.8
ENGINE_H     = 1.0
Y_ENGINE_BOT = Y_ENGINE_TOP - ENGINE_H                   # 22.8

Y_LOOP_TOP   = Y_ENGINE_BOT - GAP - 0.1                 # 22.2
# 6 boxes: each BOX_H=1.5, gap 0.5 between them
ROUND_TOTAL  = 6 * BOX_H + 5 * GAP                      # 11.5
Y_LOOP_PAD   = 0.9                                       # padding top + bottom
Y_LOOP_BOT   = Y_LOOP_TOP - ROUND_TOTAL - 2 * Y_LOOP_PAD  # 8.9 approx
LOOP_H       = Y_LOOP_TOP - Y_LOOP_BOT

# Round Y positions (top of each box, inside the loop)
Y_R = []
for i in range(6):
    Y_R.append(Y_LOOP_TOP - Y_LOOP_PAD - i * (BOX_H + GAP) - BOX_H)
# Y_R[i] is the BOTTOM-LEFT y of round i

Y_RESULT_TOP = Y_LOOP_BOT - GAP - 0.3
RESULT_H     = 1.3
Y_RESULT_BOT = Y_RESULT_TOP - RESULT_H

Y_LEGEND     = Y_RESULT_BOT - 0.8

# ═══════════════════════════════════════════════════════════
# DRAW
# ═══════════════════════════════════════════════════════════

# ── Title ──────────────────────────────────────────────────
txt(CX, Y_TITLE, 'OpsRamp Autopilot', fs=26, fw='bold', fc='#0f172a')
txt(CX, Y_SUBTITLE, 'Autonomous IT Incident Resolution powered by ReAct AI',
    fs=13, fc='#64748b', style='italic')

# ── 1. User box ───────────────────────────────────────────
uw = 8.0
ux = CX - uw / 2
box(ux, Y_USER_BOT, uw, USER_H, '#4f46e5', '#3730a3', lw=2.5)
txt(CX, Y_USER_BOT + USER_H * 0.62, 'USER QUESTION',
    fs=14, fw='bold', fc='white')
txt(CX, Y_USER_BOT + USER_H * 0.28, '"Why is the payment app slow?"',
    fs=11, fc='#c7d2fe', style='italic')

# arrow User -> Engine
arrow(CX, Y_USER_BOT - ARROW_GAP, CX, Y_ENGINE_TOP + ENGINE_H + ARROW_GAP - ENGINE_H, '#4f46e5')

# ── 2. Engine box ─────────────────────────────────────────
ew = 10.6
ex = CX - ew / 2
box(ex, Y_ENGINE_BOT, ew, ENGINE_H, '#7c3aed', '#5b21b6', lw=2.5)
txt(CX, Y_ENGINE_BOT + ENGINE_H / 2,
    'OpsRamp Autopilot  |  ReAct AI Engine  (LLM + Tool Calling)',
    fs=13, fw='bold', fc='white')

# arrow Engine -> Loop
arrow(CX, Y_ENGINE_BOT - ARROW_GAP, CX, Y_LOOP_TOP + ARROW_GAP, '#7c3aed')

# ── 3. Loop boundary ──────────────────────────────────────
loop_w = 17.0
loop_x = CX - loop_w / 2
box(loop_x, Y_LOOP_BOT, loop_w, LOOP_H, '#f8fafc', '#cbd5e1', lw=2, ls='--')
txt(CX, Y_LOOP_TOP - 0.35, 'AUTONOMOUS REASONING LOOP',
    fs=15, fw='bold', fc='#475569')

# ── 4. Round boxes (left column) ──────────────────────────
rounds = [
    ("ROUND 1   Search Alerts",
     "Finds HTTP latency spike on app-server-05",
     '#6366f1', '#eef2ff'),
    ("ROUND 2   Investigate Resource",
     "CPU 45%, Memory 60% -- both normal",
     '#6366f1', '#eef2ff'),
    ("ROUND 3   Correlate Network",
     "2% packet loss on Juniper switch SW-12",
     '#0ea5e9', '#ecfeff'),
    ("ROUND 4   Blast Radius Analysis",
     "3 apps, 24 servers, 5 000 users impacted",
     '#f59e0b', '#fffbeb'),
    ("ROUND 5   Knowledge Base (RAG)",
     "Runbook found -- restart affected interface",
     '#8b5cf6', '#f5f3ff'),
    ("ROUND 6   Guided Remediation",
     "Proposes fix --> awaits approval --> executes",
     '#10b981', '#ecfdf5'),
]

rx = LEFT_CX - BOX_W / 2   # left edge of round boxes

for i, (title, detail, bc, fc) in enumerate(rounds):
    by = Y_R[i]  # bottom-left y
    box(rx, by, BOX_W, BOX_H, fc, bc, lw=2.5)
    txt(rx + BOX_W / 2, by + BOX_H * 0.68, title,
        fs=11.5, fw='bold', fc='#1e293b')
    txt(rx + BOX_W / 2, by + BOX_H * 0.30, detail,
        fs=10, fc='#64748b', style='italic')
    # arrow to next round
    if i < 5:
        arrow(LEFT_CX, by - ARROW_GAP,
              LEFT_CX, Y_R[i + 1] + BOX_H + ARROW_GAP, bc)

# ── 5. Data source boxes (right column) ───────────────────
txt(RIGHT_CX, Y_LOOP_TOP - 0.35, 'DATA SOURCES',
    fs=14, fw='bold', fc='#475569')

# Each source aligned vertically to its connected round
src_sx = RIGHT_CX - SRC_W / 2

sources = [
    ("OpsRamp API\nAlerts | Resources\nIncidents | Metrics",
     '#eef2ff', '#6366f1', 0),    # aligned to round 0
    ("Juniper API\nTopology | Ports\nTelemetry",
     '#ecfeff', '#0ea5e9', 2),    # aligned to round 2
    ("Dependency Graph\nNetwork <> Server\n<> Application",
     '#fffbeb', '#f59e0b', 3),    # aligned to round 3
    ("Runbook PDFs\nRAG Vector Store",
     '#f5f3ff', '#8b5cf6', 4),    # aligned to round 4
]

for label, fc, bc, ri in sources:
    # center source box vertically at same height as the connected round
    round_cy = Y_R[ri] + BOX_H / 2
    sy = round_cy - SRC_H / 2
    box(src_sx, sy, SRC_W, SRC_H, fc, bc, lw=2)
    txt(RIGHT_CX, round_cy, label,
        fs=9.5, fw='bold', fc='#334155', linespacing=1.3)
    # dashed line from right edge of round to left edge of source
    dash(rx + BOX_W, round_cy, src_sx, round_cy, bc)

# ── 6. Arrow last round -> result ─────────────────────────
arrow(LEFT_CX, Y_R[5] - ARROW_GAP,
      LEFT_CX, Y_RESULT_TOP + ARROW_GAP - RESULT_H + RESULT_H, '#10b981')

# ── 7. Resolution box (full width, centered) ──────────────
res_w = 15.0
res_x = CX - res_w / 2
box(res_x, Y_RESULT_BOT, res_w, RESULT_H, '#059669', '#047857', lw=2.5)
txt(CX, Y_RESULT_BOT + RESULT_H * 0.68, 'RESOLUTION',
    fs=16, fw='bold', fc='white')
txt(CX, Y_RESULT_BOT + RESULT_H * 0.30,
    'Root Cause: Juniper SW-12 packet loss   |   Impact: 3 apps, 5000 users   |   Fix: Interface restart (approved)',
    fs=10, fc='#d1fae5')

# ── 8. Legend ──────────────────────────────────────────────
# Solid arrow legend
lx1 = 3.5
ax.annotate('', xy=(lx1 + 1.8, Y_LEGEND), xytext=(lx1, Y_LEGEND),
            arrowprops=dict(arrowstyle='->', color='#64748b', lw=2))
txt(lx1 + 2.3, Y_LEGEND, 'Reasoning flow', fs=10, fc='#64748b', ha='left')

# Dashed line legend
lx2 = 9.0
ax.plot([lx2, lx2 + 1.5], [Y_LEGEND, Y_LEGEND],
        linestyle=(0, (6, 4)), color='#94a3b8', lw=1.6)
txt(lx2 + 2.0, Y_LEGEND, 'Data lookup', fs=10, fc='#94a3b8', ha='left')

# ── Save ───────────────────────────────────────────────────
plt.subplots_adjust(left=0.02, right=0.98, top=0.98, bottom=0.02)
out = '/Users/a60168034/go/src/ai/opsRampChatBot/conversationalAgent/screenshots/opsramp_autopilot_flow.png'
plt.savefig(out, dpi=200, bbox_inches='tight', facecolor='white', edgecolor='none')
print(f"Saved: {out}")
