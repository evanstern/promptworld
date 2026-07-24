# cast_light: a scripted omen that branches on the world clock (spec 036 US3,
# contracts/script-api.md example). At night the light blooms for every living
# villager; by day it is invisible and only its target notices nothing. It
# declares only agent.memory_added and charges 0 — a narrate-only tool produces
# perception memories, which the reducer does not price, so a nonzero charge gate
# would misrepresent the cost.

def apply(args, world):
    if world.time_of_day == "night":
        return [
            {"kind": "narrate",
             "text": "A soft light blooms over " + args["target"] + ".",
             "recipients": "all_living"},
        ]
    return [
        {"kind": "narrate",
         "text": "The light is invisible in daylight.",
         "recipients": "target",
         "target": args["target"]},
    ]
