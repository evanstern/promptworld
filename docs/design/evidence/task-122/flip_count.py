#!/usr/bin/env python3
"""TASK-122 flip-rate measurement — the SC-007 method, verbatim, plus warm_up.

Usage: python3 flip_count.py <world.db path>
Read-only. Classes per the TASK-101 spike / SC-007 evidence:
  food   = {forage, hunt}
  warmth = {goto_warmth, build_fire, refuel_fire, warm_up}   # warm_up: spec 064 verb
A flip = consecutive classified intent_set records (per agent) changing class.
Reports total flips, <=200-tick flips, flips/game-day (86400 ticks/day).
Bar (SC-007): worst agent <= 36 flips/game-day. Baseline worst: Sage ~72/day.
"""
import json, sqlite3, sys
FOOD = {"forage", "hunt"}
WARMTH = {"goto_warmth", "build_fire", "refuel_fire", "warm_up"}
def main(db):
    con = sqlite3.connect(f"file:{db}?mode=ro", uri=True)
    rows = con.execute("SELECT tick, payload FROM events WHERE type='agent.intent_set' ORDER BY seq")
    last, flips, fast, counts = {}, {}, {}, {}
    lo = hi = None
    for tick, payload in rows:
        d = json.loads(payload)
        a, g = d.get("agent"), d.get("goal")
        cls = "food" if g in FOOD else "warmth" if g in WARMTH else None
        lo = tick if lo is None else lo
        hi = tick
        if cls is None: continue
        counts[a] = counts.get(a, 0) + 1
        if a in last and last[a][0] != cls:
            flips[a] = flips.get(a, 0) + 1
            if tick - last[a][1] <= 200: fast[a] = fast.get(a, 0) + 1
        last[a] = (cls, tick)
    days = (hi - lo) / 86400 if hi and lo is not None else 0
    print(f"span: ticks {lo}..{hi} = {days:.3f} game-days")
    worst = 0.0
    for a in sorted(set(counts) | set(flips)):
        rate = (flips.get(a, 0) / days) if days else 0
        worst = max(worst, rate)
        print(f"agent {a}: flips={flips.get(a,0)} fast(<=200t)={fast.get(a,0)} rate={rate:.2f}/day n={counts.get(a,0)}")
    print(f"WORST: {worst:.2f}/day (bar <=36; baseline worst ~72)")
if __name__ == "__main__":
    main(sys.argv[1])
