def apply(args, world):
    # Script mode composes the SAME spec-082 address grammar as declarative
    # targets — one shared compile path, no script-specific surface (FR-007).
    return [{"kind": "remove_entity", "target": "pile@%d,%d" % (args["x"], args["y"])}]
