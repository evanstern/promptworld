#!/usr/bin/env python3
"""TASK-106 thrash-detection research on world-01 v3 event log (read-only).

Ground truth DB: ~/.promptworld/worlds/world-01/world.v3.db
Ticks are game-seconds: 3600 ticks = 1 game-hour, 86400 = 1 game-day.
Agents: 0=Ash 1=Birch 2=Cedar 3=Rowan 4=Fern 5=Hazel 6=Oak 7=Sage
"""
import sqlite3, json, bisect, math, sys
from collections import defaultdict

DB = "/Users/evanstern/.promptworld/worlds/world-01/world.v3.db"
OUT = "/Users/evanstern/.claude/jobs/8a0b1593/tmp/task106"
NAMES = ["Ash", "Birch", "Cedar", "Rowan", "Fern", "Hazel", "Oak", "Sage"]
HOUR = 3600
FOOD_CLASS = {"forage", "hunt", "eat"}
WARMTH_CLASS = {"goto_warmth", "build_fire", "refuel_fire"}
EPS = 5  # need-improvement epsilon on the 0..1000 scale

# broad classes for switch-rate metric
BROAD = {}
for g in FOOD_CLASS: BROAD[g] = "food"
for g in WARMTH_CLASS: BROAD[g] = "warmth"
for g in ("cook",): BROAD[g] = "food"
for g in ("sleep",): BROAD[g] = "rest"
for g in ("collect_water", "bathe"): BROAD[g] = "water"
for g in ("chop", "quarry", "craft_stone", "craft_planks", "craft_spear",
          "build_oven", "build_chest", "build_wall_stone", "build_path",
          "drop", "pick_up", "deposit"): BROAD[g] = "work"
for g in ("seek", "attend_meeting"): BROAD[g] = "social"
for g in ("wander",): BROAD[g] = "idle"

def clazz(goal):
    if goal in FOOD_CLASS: return "F"
    if goal in WARMTH_CLASS: return "W"
    return None

def gtime(t):
    """Tick 0 = day 1 06:00 (sim.day_started fires at ticks ≡ 0 mod 86400,
    sim.night_started at +57600 → 16h day, 8h night)."""
    abs_s = t + 6 * HOUR
    d = abs_s // 86400 + 1
    r = abs_s % 86400
    return f"day{d} {r//3600:02d}:{(r%3600)//60:02d}"

con = sqlite3.connect(f"file:{DB}?mode=ro", uri=True)

# ---- load intents ----
intents = defaultdict(list)  # agent -> [(tick, goal, source, tx, ty)]
for tick, payload in con.execute(
        "SELECT tick, payload FROM events WHERE type='agent.intent_set' ORDER BY seq"):
    p = json.loads(payload)
    intents[p["agent"]].append((tick, p["goal"], p.get("source", "?"),
                                p.get("target_x"), p.get("target_y")))

# ---- load needs (food, warmth) ----
needs_t = defaultdict(list); needs_f = defaultdict(list); needs_w = defaultdict(list)
for tick, payload in con.execute(
        "SELECT tick, payload FROM events WHERE type='agent.needs_changed' ORDER BY seq"):
    p = json.loads(payload)
    a = p["agent"]
    needs_t[a].append(tick); needs_f[a].append(p["food"]); needs_w[a].append(p["warmth"])

def need_at(a, tick):
    """(food, warmth) as of last sample <= tick; None if before first sample."""
    i = bisect.bisect_right(needs_t[a], tick) - 1
    if i < 0: return None
    return needs_f[a][i], needs_w[a][i]

# ---- load positions (for wasted travel) ----
pos_t = defaultdict(list); pos_xy = defaultdict(list)
for tick, payload in con.execute(
        "SELECT tick, payload FROM events WHERE type='agent.moved' ORDER BY seq"):
    p = json.loads(payload)
    pos_t[p["agent"]].append(tick); pos_xy[p["agent"]].append((p["x"], p["y"]))

def walked(a, t0, t1):
    """Actual path length (Euclidean per step) from agent.moved in [t0,t1]."""
    i = bisect.bisect_left(pos_t[a], t0); j = bisect.bisect_right(pos_t[a], t1)
    seg = pos_xy[a][max(0, i - 1):j]
    return sum(math.dist(seg[k], seg[k + 1]) for k in range(len(seg) - 1))

death = {0: 101040, 6: 511440}  # Ash starvation, Oak exposure
MAXTICK = 538823

# =========== 2. per-agent intent history ===========
report2 = {}
for a in range(8):
    seq = intents[a]
    bygoal = defaultdict(int); bysrc = defaultdict(int)
    for _, g, s, _, _ in seq:
        bygoal[g] += 1; bysrc[s] += 1
    # class subsequence (only F/W intents)
    sub = [(t, clazz(g), g, tx, ty) for t, g, s, tx, ty in seq if clazz(g)]
    flips_sub = sum(1 for i in range(1, len(sub)) if sub[i][1] != sub[i - 1][1])
    # strict-adjacent flips (consecutive in the FULL stream, classes differ)
    flips_strict = 0
    for i in range(1, len(seq)):
        c0, c1 = clazz(seq[i - 1][1]), clazz(seq[i][1])
        if c0 and c1 and c0 != c1: flips_strict += 1
    report2[a] = dict(name=NAMES[a], total=len(seq), by_goal=dict(bygoal),
                      by_source=dict(bysrc), flips_subseq=flips_sub,
                      flips_strict=flips_strict, class_intents=len(sub))

# =========== 3. episode identification ===========
# flip stream per agent: tick of the LATER intent of each flipping pair (subsequence def)
flipticks = {}
for a in range(8):
    sub = [(t, clazz(g)) for t, g, s, _, _ in intents[a] if clazz(g)]
    flipticks[a] = [sub[i][0] for i in range(1, len(sub)) if sub[i][1] != sub[i - 1][1]]

def find_episodes(a, gap=2 * HOUR, min_flips=5):
    eps_ = []
    cur = []
    for t in flipticks[a]:
        if cur and t - cur[-1] > gap:
            if len(cur) >= min_flips: eps_.append(cur)
            cur = []
        cur.append(t)
    if len(cur) >= min_flips: eps_.append(cur)
    out = []
    for c in eps_:
        t0, t1 = c[0], c[-1]
        dur_h = max((t1 - t0) / HOUR, 1 / 60)
        n0, n1 = need_at(a, t0), need_at(a, t1)
        out.append(dict(t0=t0, t1=t1, t0h=gtime(t0), t1h=gtime(t1),
                        flips=len(c), flips_per_h=round(len(c) / dur_h, 1),
                        dur_h=round(dur_h, 2),
                        food_delta=(n1[0] - n0[0]) if n0 and n1 else None,
                        warmth_delta=(n1[1] - n0[1]) if n0 and n1 else None,
                        food_end=n1[0] if n1 else None, warmth_end=n1[1] if n1 else None))
    return out

episodes = {a: find_episodes(a) for a in range(8)}

# known-bad = every episode with >=20 flips on Sage/Fern/Oak (plus all listed)
known_bad = []
for a in (7, 4, 6):
    for e in episodes[a]:
        if e["flips"] >= 20:
            known_bad.append((a, e))

# =========== 4. detector sweep ===========
def detect(a, W, K, need_clause):
    """Sliding window ending at each class-intent. ABA count = flips_in_window - 1.
    Fires when ABA >= K (and need clause passes, if enabled).
    Returns (firing_ticks, merged_episodes [(t0,t1)])."""
    sub = [(t, clazz(g)) for t, g, s, _, _ in intents[a] if clazz(g)]
    fires = []
    for i in range(len(sub)):
        t = sub[i][0]
        lo = t - W
        j = i
        while j > 0 and sub[j - 1][0] >= lo: j -= 1
        win = sub[j:i + 1]
        flips = sum(1 for k in range(1, len(win)) if win[k][1] != win[k - 1][1])
        aba = max(0, flips - 1)
        if aba < K: continue
        if need_clause:
            n0, n1 = need_at(a, lo), need_at(a, t)
            if n0 is None or n1 is None: continue
            food_impr = n1[0] - n0[0] > EPS
            warm_impr = n1[1] - n0[1] > EPS
            if food_impr or warm_impr: continue  # some need improved -> healthy
        fires.append(t)
    # merge fires closer than W into episodes
    merged = []
    for t in fires:
        if merged and t - merged[-1][1] <= W: merged[-1][1] = t
        else: merged.append([t, t])
    return fires, merged

def overlaps(ep, dets, W):
    return any(d0 - W <= ep["t1"] and d1 >= ep["t0"] for d0, d1 in dets)

sweep = {}
for W in (2 * HOUR, 4 * HOUR, 8 * HOUR):
    for K in (3, 5, 8):
        cell = dict(per_agent={}, bad_caught={}, )
        for a in range(8):
            f_no, m_no = detect(a, W, K, False)
            f_yes, m_yes = detect(a, W, K, True)
            cell["per_agent"][NAMES[a]] = dict(
                fires_raw=len(f_no), episodes_raw=len(m_no),
                fires_clause=len(f_yes), episodes_clause=len(m_yes))
            cell.setdefault("_dets", {})[a] = (m_no, m_yes)
        for a, e in known_bad:
            key = f"{NAMES[a]}@{e['t0h']}"
            m_no, m_yes = cell["_dets"][a]
            cell["bad_caught"][key] = dict(raw=overlaps(e, m_no, W),
                                           clause=overlaps(e, m_yes, W))
        del cell["_dets"]
        sweep[f"W={W // HOUR}h,K={K}"] = cell

# =========== healthy-interleave false positives ===========
# raw detections whose window shows BOTH needs >=500 at fire time AND at least one
# need improving -> "healthy interleaving"; count how many the clause removes.
def healthy_fp(a, W, K):
    sub = [(t, clazz(g)) for t, g, s, _, _ in intents[a] if clazz(g)]
    fp = 0; removed = 0
    for i in range(len(sub)):
        t = sub[i][0]; lo = t - W
        j = i
        while j > 0 and sub[j - 1][0] >= lo: j -= 1
        win = sub[j:i + 1]
        flips = sum(1 for k in range(1, len(win)) if win[k][1] != win[k - 1][1])
        if max(0, flips - 1) < K: continue
        n0, n1 = need_at(a, lo), need_at(a, t)
        if not n0 or not n1: continue
        healthy = min(n1) >= 500 and (n1[0] - n0[0] > EPS or n1[1] - n0[1] > EPS)
        if healthy:
            fp += 1
            if n1[0] - n0[0] > EPS or n1[1] - n0[1] > EPS: removed += 1
    return fp, removed

fp_table = {}
for W in (2 * HOUR, 4 * HOUR, 8 * HOUR):
    for K in (3, 5, 8):
        tot_fp = 0
        for a in range(8):
            fp, rem = healthy_fp(a, W, K)
            tot_fp += fp
        fp_table[f"W={W // HOUR}h,K={K}"] = tot_fp

# =========== 5a. switch-rate (all classes) ===========
switch = {}
for a in range(8):
    seq = [(t, BROAD.get(g, g)) for t, g, s, _, _ in intents[a]]
    end = death.get(a, MAXTICK)
    hours = end / HOUR
    ch = sum(1 for i in range(1, len(seq)) if seq[i][1] != seq[i - 1][1])
    switch[NAMES[a]] = dict(class_changes=ch, alive_hours=round(hours, 1),
                            per_hour=round(ch / hours, 2))

# =========== 5b. wasted-travel ratio on known-bad episodes + healthy baseline ===========
travel = []
for a, e in known_bad:
    dist = walked(a, e["t0"], e["t1"])
    gain = max(0, e["food_delta"] or 0) + max(0, e["warmth_delta"] or 0)
    travel.append(dict(agent=NAMES[a], t0h=e["t0h"], t1h=e["t1h"],
                       walked_tiles=round(dist, 0), need_gain=gain,
                       ratio=round(dist / (gain + 1), 2)))
# healthy baseline: Cedar & Rowan over an arbitrary healthy 8h window day 4 noon
for a in (2, 3):
    t0, t1 = 4 * 86400 + 8 * HOUR, 4 * 86400 + 16 * HOUR
    n0, n1 = need_at(a, t0), need_at(a, t1)
    dist = walked(a, t0, t1)
    gain = max(0, n1[0] - n0[0]) + max(0, n1[1] - n0[1])
    travel.append(dict(agent=NAMES[a] + " (healthy baseline)", t0h=gtime(t0), t1h=gtime(t1),
                       walked_tiles=round(dist, 0), need_gain=gain,
                       ratio=round(dist / (gain + 1), 2)))

out = dict(report2=report2, episodes={NAMES[a]: episodes[a] for a in range(8)},
           known_bad=[(NAMES[a], e) for a, e in known_bad],
           sweep=sweep, fp_table=fp_table, switch=switch, travel=travel)
with open(f"{OUT}/raw_results.json", "w") as f:
    json.dump(out, f, indent=1, default=str)
print("written", f"{OUT}/raw_results.json")
