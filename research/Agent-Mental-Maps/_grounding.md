---
title: Agent Mental Maps — Grounding
aliases: []
tags: [grounding]
type: source
created: 2026-07-24
updated: 2026-07-24
related: ["[[Agent-Mental-Maps]]"]
---

# Agent Mental Maps — Grounding

> Source-of-truth artifact. This is the raw, cited output of a research pass (the `deep-research`
> skill, or a direct web-search fan-out). Keep it close to verbatim — do not editorialize, prune,
> or draw conclusions here. Knowledge notes and analyses cite *into* this file.

**Research question:** How can agents build and maintain "mental maps" of 2D grid worlds
(64x64 tiles), with special attention to algorithms and data structures that extend to
3 dimensions (layered grids / voxel stacks)? Covering: hierarchical spatial data structures,
cognitive-map / spatial-memory models from game AI and robotics, fog-of-war / knowledge-limited
perception, belief decay and stale-knowledge handling, hierarchical pathfinding over known
space, exploration of unknown space, multi-agent map sharing, and memory-efficient per-agent
representations.
**Method:** web-search fan-out (11 parallel searches + 1 reference fetch) · 2026-07-24

---

## 1. Hierarchical spatial tree structures (quadtrees, octrees, k-d trees, BVHs)

- A **quadtree** is a hierarchical tree data structure that recursively divides 2D space into
  four bounding squares (quadrants); it was introduced roughly four decades ago and is the
  canonical hierarchical structure for 2D spatial data
  ([Medium — Quadtree](https://medium.com/@yeshsurya/quadtree-a-spatial-data-structure-for-efficient-queries-f4f92958881d),
  [Samet, "The Quadtree and Related Hierarchical Data Structures", CSUR 1984](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)).
- An **octree** is the three-dimensional analog of the quadtree: each internal node has exactly
  eight children, dividing space into bounding cubes instead of bounding squares
  ([GameDev.net — Introduction to Octrees](https://www.gamedev.net/articles/programming/general-and-gameplay-programming/introduction-to-octrees-r3529/)).
- Usage guidance from game-development practice: quadtrees suit "sprawling topology that is
  roughly-2D in nature, like terrain, maybe with items scattered over it"; octrees suit scenes
  that "spread out in all directions around you (up, down, in front, behind, left and right)"
  ([GameDev.net forum — Using octrees for spatial representation](https://www.gamedev.net/forums/topic/666717-using-octrees-for-spatial-representation/)).
- A documented middle path for mostly-2D worlds with some verticality: keep a quadtree but add
  "additional information or layers … to the leaf nodes to specify 3D aspects of the game
  world," acceptable "when the additional dimension (e.g., height) is not as detailed or
  complex as the other dimensions"
  ([GameDev.net forum — Using QuadTrees in maps](https://gamedev.net/forums/topic/463418-using-quadtrees-in-maps/)).
- Quadtrees are widely used in games for collision detection and spatial queries; beyond games
  they appear in image processing, computer graphics, GIS, and robotics
  ([Medium — Quadtree](https://medium.com/@yeshsurya/quadtree-a-spatial-data-structure-for-efficient-queries-f4f92958881d),
  [Samet 1984](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)).
- Quadtrees "extend the occupancy grid concept by hierarchically subdividing non-uniform
  regions into quadrants while merging homogeneous areas, providing a more compact and flexible
  spatial representation"; **probabilistic quadtrees** extend this to efficiently represent
  probabilistic occupancy maps
  ([arXiv — 2D Visibility in Cartesian Grid Maps](https://arxiv.org/pdf/2403.06494)).
- Region quadtrees achieve "compact representations of large two-dimensional binary arrays" —
  homogeneous regions collapse to single leaves
  ([arXiv — 2D Visibility in Cartesian Grid Maps](https://arxiv.org/pdf/2403.06494),
  [Samet 1984](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)).

## 2. Occupancy grids and probabilistic belief maps (robotics)

- **Occupancy grid mapping** represents the environment as a grid (2D or 3D) of cells, each
  holding a probability of being occupied; it is "a foundational framework in robotic
  perception," discretizing space and refining per-cell occupancy with a **binary Bayes
  filter** ([Medium — Occupancy Grid Mapping Algorithm](https://medium.com/@SuriNaren/occupancy-grid-mapping-algorithm-e451701da0e8),
  [Freiburg robotics course slides](http://ais.informatik.uni-freiburg.de/teaching/ss16/robotics/slides/12-occupancy-mapping.pdf)).
- Cells store **log-odds**, not raw probabilities: each observation adds a positive value
  (occupied evidence) or negative value (free evidence). "A cell with a positive log odds is
  considered as occupied, negative is considered as free, and zero is considered as unknown
  (equivalent to a probability of 0.5, indicating no observation for the cell yet)"
  ([CMU 16-831 lecture notes](https://www.cs.cmu.edu/~16831-f12/notes/F12/16831_lecture05_vh.pdf),
  [ThinkAutonomous — Occupancy Grid Mapping](https://www.thinkautonomous.ai/blog/occupancy-grid-mapping/)).
- Why log-odds: it "factorizes a high-dimensional mapping problem into a series of simple
  recursive additions"; updates are integer/float additions ("a robot can update millions of
  cells per second"), and values stay in a manageable numeric range, avoiding the underflow of
  repeated probability multiplication
  ([ThinkAutonomous](https://www.thinkautonomous.ai/blog/occupancy-grid-mapping/),
  [CMU 16-831](https://www.cs.cmu.edu/~16831-f12/notes/F12/16831_lecture05_vh.pdf)).
- The standard occupancy map is **tri-state in effect**: occupied / free / unknown, with
  unknown as the prior (0.5). Confidence-aware variants additionally track how much evidence
  supports each cell
  ([Confidence-aware Occupancy Grids, IROS WS](https://karolhausman.github.io/pdf/agha17-ws-iros.pdf)).
- **Time-aware occupancy grid mapping** exists as a patented technique for robots in dynamic
  environments — occupancy estimates account for the age of observations
  ([USPTO patent 12,523,498](https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/12523498)).

## 3. OctoMap — the probabilistic octree (the standard 3D extension)

- **OctoMap** is "an efficient probabilistic 3D mapping framework based on octrees": 3D space
  is discretized into equal-sized voxels stored in an octree where every internal node divides
  into 8 children; it explicitly represents **occupied, free, and unknown** space
  ([Hornung et al., Autonomous Robots 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0),
  [ACM DL](https://dl.acm.org/doi/10.1007/s10514-012-9321-0)).
- Each leaf stores the **log-odds of occupancy** — i.e., OctoMap is exactly the occupancy-grid
  belief model (section 2) lifted into an octree, showing the 2D↔3D continuity of the
  technique ([Hornung et al. 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0)).
- OctoMap includes an octree **map compression** method keeping 3D models compact (merging
  identical children), ships as an open-source C++ library, and is very widely used in
  robotics (~3,000+ citations)
  ([Hornung et al. 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0),
  [SciSpace citation record](https://scispace.com/papers/octomap-an-efficient-probabilistic-3d-mapping-framework-3dp42nnofi)).
- Memory-lean alternatives to voxel/octree maps exist for constrained platforms, e.g. **GMMap**
  (Gaussian-mixture continuous occupancy) which "accurately construct[s] and quer[ies] maps
  while preserving unexplored regions to reduce memory overhead"
  ([arXiv — GMMap](https://arxiv.org/pdf/2306.03740)); non-uniform cell representations for
  dynamic occupancy grids likewise trade resolution for memory
  ([ResearchGate — non-uniform cell occupancy mapping](https://www.researchgate.net/publication/348367869_Efficient_dynamic_occupancy_grid_mapping_using_non-uniform_cell_representation)).

## 4. Cognitive maps, topological maps, and landmark graphs

- Cognitive-science framing carried into robotics/AI: "humans build approximate graphs of
  their environment, encoding relative distances between landmarks"; in vision and robotics
  "these ideas have translated to the construction of topological graphs and latent maps"
  ([arXiv — Memory Proxy Maps for Visual Navigation](https://arxiv.org/html/2411.09893)).
- "Cognitive research suggests that humans save landmark features in their memory for
  navigation, instead of detailed scene layouts, and several methods utilize this theory to
  propose topological memory for visual navigation"
  ([MemoNav](https://arxiv.org/pdf/2208.09610), [MemoNav v2](https://arxiv.org/html/2402.19161v2)).
- For embodied navigation, "memory should organize online perception into a relational spatial
  state that summarizes explored regions, traversed paths, landmarks, and their spatial
  connectivity" ([SpaceVLN](https://arxiv.org/pdf/2606.08992)).
- SpaceVLN builds a hierarchical spatial memory online: a **spatial waypoint graph** plus a
  **local landmark memory**, shared between planner and executor
  ([SpaceVLN](https://arxiv.org/pdf/2606.08992)).
- Survey-level position: "future research should focus on developing **hybrid cognitive map
  architectures that unify topological, metric, and semantic representations** into a flexible
  and context-aware spatial memory system"
  ([Frontiers in Computational Neuroscience — dynamic cognitive maps](https://www.frontiersin.org/journals/computational-neuroscience/articles/10.3389/fncom.2024.1498160/full),
  [arXiv — Mind Meets Space](https://arxiv.org/pdf/2509.09154)).
- Robot navigation systems using "topological memory configuration" pair a learned local
  controller with a sparse graph of previously visited places
  ([IEEE/CAA JAS — Cognitive Navigation](https://www.ieee-jas.net/article/doi/10.1109/JAS.2024.124332));
  semantic-spatial retrieval systems (e.g. Meta-Memory) integrate semantic and spatial memories
  for spatial reasoning ([arXiv — Meta-Memory](https://arxiv.org/html/2509.20754v1)).

## 5. Fog of war and knowledge-limited perception in games/simulations

- **Fog of war** is the standard game mechanic "to simulate limited knowledge or visibility of
  the game world"; it "adds strategic depth by requiring players to gather information, scout,
  and make decisions based on partial knowledge, creating uncertainty and encouraging
  exploration" ([Machinations.io glossary](https://machinations.io/glossary/fog-of-war)).
- The common tile-based implementation keeps a per-player **visibility grid with three states**
  — unexplored (never seen), explored-but-not-currently-visible (terrain remembered, dynamic
  entities hidden/stale), and currently visible — updated from unit sight radii each tick
  ([Didac Romero — tile-based Fog of War in SDL/C++ tutorial](https://didacromero.github.io/Fog-of-War/)).
- In RTS AI research, fog of war makes the environment **partially observable**: "terrain is
  partly unknown and must be explored; bots must be able to update their knowledge about
  terrain and have units explore unknown terrain to locate hidden enemy bases"
  ([Dealing with Fog of War in a RTS Game Environment](https://www.researchgate.net/publication/224491323_Dealing_with_Fog_of_War_in_a_Real_Time_Strategy_Game_Environment)).
- Agent designs under fog of war handle uncertainty explicitly — e.g. Bayesian networks with
  parameters learned from game logs, or encoder-decoder models predicting hidden enemy units
  from partial, noisy observations
  ([StarCraft fog-of-war prediction study](https://www.researchgate.net/publication/313279929_Investigation_of_the_Effect_of_Fog_of_War_in_the_Prediction_of_StarCraft_Strategy_Using_Machine_Learning),
  [Dealing with Fog of War](https://www.academia.edu/79349356/Dealing_with_fog_of_war_in_a_Real_Time_Strategy_game_environment)).
- Multi-agent programming contest agents "maintain internal environment representations using
  visibility grids initialized with values denoting unobserved cells, with these grids
  recording what parts of the environment have been explored and what information has been
  observed" ([arXiv — Requirement Gatherers, MAPC 2019](https://arxiv.org/pdf/2006.02816)).

## 6. Belief decay, staleness, and provenance

- In belief-based agent memory, "unused beliefs **decay toward uncertainty**, and the decay
  rate controls the agent's reliance on historical memory, influencing the balance between
  exploitation of seen patterns and robust generalization"
  ([arXiv — Belief Memory: Agent Memory Under Partial Observability](https://arxiv.org/html/2605.05583v1)).
- Stale memories — "facts that were once true but are no longer valid" — have an **asymmetric
  failure mode**: "stale memory rarely prevents retrieval, but regularly leads the agent to act
  confidently on invalidated assumptions." Recommended handling: weight retrieval by a
  **staleness penalty** (time since last verification) and a confidence-gated risk term, and
  "treat the retrieved content as a hypothesis until re-checked against the live environment"
  ([Zylos Research — AI Agent Memory Architectures](https://zylos.ai/research/2026-04-05-ai-agent-memory-architectures-persistent-knowledge/)).
- Memory maintenance in dynamic environments must "track what changed, what stayed true, what
  was contradicted, what should decay, what should become stable knowledge, and what should
  remain as raw evidence"
  ([Zylos Research](https://zylos.ai/research/2026-04-05-ai-agent-memory-architectures-persistent-knowledge/)).
- **BeliefMem** shifts agent memory "from storing deterministic conclusions to maintaining an
  attribute-level belief representation over the environment," keeping multiple candidate
  conclusions per fact with probabilities "updated via noisy-OR evidence merge as new
  observations arrive"; reliability-conditional updating also serves as a defense against
  poisoned/unreliable reports
  ([arXiv — When Does Belief-Based Agent Memory Help?](https://arxiv.org/html/2606.22030)).
- Robotics parallel: time-aware occupancy grids age out observations in dynamic environments
  ([USPTO 12,523,498](https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/12523498));
  confidence-aware occupancy grids track evidence mass per cell
  ([Confidence-aware Occupancy Grids](https://karolhausman.github.io/pdf/agha17-ws-iros.pdf)).

## 7. Hierarchical pathfinding over grid maps (HPA*, quadtree A*)

- **HPA\*** (Hierarchical Path-Finding A\*, Botea et al. 2004) "abstracts a map into linked
  local clusters, where at the local level the optimal distances for crossing each cluster are
  pre-computed and cached, and at the global level clusters are traversed in a single big
  step"; it finds paths **within 1% of optimal**
  ([Botea & Müller — Near Optimal Hierarchical Path-Finding](https://www.researchgate.net/publication/228785110_Near_optimal_hierarchical_path-finding_HPA),
  [Semantic Scholar record](https://www.semanticscholar.org/paper/Near-Optimal-Hierarchical-Path-Finding-Botea-M%C3%BCller/b0f0432ba69e4d730b93a75e3d19c8e9d811efac)).
- The hierarchy "can be extended to more than two levels, where small clusters are grouped
  together to form larger clusters"; crossing distances for a large cluster reuse those of the
  contained small clusters ([Botea & Müller](https://citeseerx.ist.psu.edu/document?doi=b0f0432ba69e4d730b93a75e3d19c8e9d811efac)).
- HPA\* is "considerably faster than A\* in all cases," at the cost of a pre-processing phase
  that grows with the number of abstraction levels
  ([Analysis of A\* vs HPA\* efficiency](https://www.researchgate.net/publication/272551633_The_Analysis_of_Efficiency_Dependence_of_the_Shortest_Path_Finding_Algorithms_A_and_HPA));
  a maintained open-source implementation is tested on Dragon Age: Origins maps
  ([hugoscurti/hierarchical-pathfinding](https://github.com/hugoscurti/hierarchical-pathfinding)).
- **Quadtree pathfinding**: the quadtree "splits the grid map into multiple sections, where a
  section … contains no obstacles or all obstacles. Adjacent quadtree nodes are connected by
  multiple gates" (a gate = two adjacent cells, one per side, directed); A\*/flow-fields run
  over the leaf adjacency graph ([hit9/QuadtreePathfinding](https://github.com/hit9/QuadtreePathfinding)).
- Reported performance: "finding a path with an A\* algorithm on a quadtree needs only 2% of
  the time" compared with a regular grid on the same indoor-environment benchmark (0.75 s →
  ~0.015 s) ([Using Quadtrees for Realtime Pathfinding in Indoor Environments](https://www.researchgate.net/publication/235443232_Using_Quadtrees_for_Realtime_Pathfinding_in_Indoor_Environments)).
- Hierarchical global A\* on irregular grids "searches for a path in a coarse initial irregular
  grid structure then proceeds with the search in refined regions of interest where obstacles
  are found" ([ScienceDirect — hierarchical pathfinding on large terrains](https://www.sciencedirect.com/science/article/abs/pii/S0957417419305081)).
- General map-representation guidance (Amit Patel): "the fewer nodes in your map
  representation, the faster A\* will be"; pathfinding cost scales worse than linearly with
  distance (doubling distance roughly quadruples the search area); hierarchical/multi-level
  representations "handle both distant and local pathfinding by ignoring unnecessary details
  at coarser levels … though sacrificing some path optimality." Grids win when movement costs
  vary per-tile; sparser graph representations win when costs are uniform. Recommendation:
  "start with your existing game world representation, then optimize if needed"
  ([Amit Patel — Map Representations](http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html)).

## 8. Exploration of unknown space (frontier-based)

- **Frontiers** are "boundaries between the explored and unexplored space" — formally, free
  cells adjacent to at least one unknown cell. **Frontier-based exploration** (Yamauchi's
  classic method) repeatedly detects frontiers and moves toward them "until there are no more
  frontiers and therefore no more unknown regions"
  ([Frontiers in Robotics & AI — detecting frontier cells](https://www.frontiersin.org/journals/robotics-and-ai/articles/10.3389/frobt.2021.616470/full),
  [Awabot — frontier exploration](https://awabot.com/en/autonomous-exploration-method-frontiers/)).
- The algorithm pairs naturally with occupancy grids: scan the grid for free cells adjacent to
  unknown cells, cluster contiguous frontier cells into contours (often via BFS), pick a
  frontier, navigate to it, update the map, repeat
  ([Topiwala — Frontier Based Exploration](https://arxiv.org/pdf/1806.03581),
  [Gu & Xu — Frontier Based Exploration for Map Building](https://cs.unh.edu/~tg1034/project/TianyiGu_AutonomousMapping.pdf)).
- Variants add semantics/topology (topological frontier-based exploration) or
  uncertainty-aware information prediction to pick the most informative frontier
  ([PMC — Topological Frontier-Based Exploration](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC6832505/),
  [arXiv — Uncertainty-Aware Information Prediction](https://arxiv.org/pdf/2412.12825)).

## 9. Multi-level / 3D navigation (layered grids, stairs, floors)

- Game-industry practice for multi-floor pathfinding (Software Inc. dev): a **layered
  pathfinding algorithm** — "rather than measuring distance geometrically, this approach asks
  'what is the shortest amount of doors/elevators to traverse to get from room A to room B?'"
  — i.e., a topological layer over per-floor grids
  ([ModDB — Pathfinding on multiple floors](https://www.moddb.com/games/software-inc/features/pathfinding-on-multiple-floors)).
- Robotics multi-floor navigation makes the same move: "transitions between floors don't
  require highly accurate 3D maps since they are typically connected only by elevators or
  stairs, so environments can be modeled as a **floor-stair topology**"
  ([Multi-Floor Zero-Shot Object Navigation](https://arxiv.org/pdf/2409.10906)).
- Stair handling: "detect stair regions and represent them as traversable areas, allowing the
  planner to generate paths over stairs"; inter-floor navigation "designates the stair area as
  the next waypoint" ([Stairway to Success](https://arxiv.org/pdf/2505.23019)).
- **2.5D elevation maps** (grid + continuous height per cell) represent terrain but are
  "unsuitable for scenarios with overhanging objects or multi-floor structures" — the boundary
  where layered/voxel representations become necessary
  ([Point Cloud Tomography](https://arxiv.org/pdf/2403.07631)).
- Grid-based multi-level A\* in games commonly builds node grids per level/plane and links
  them at stair nodes ([3D-A-Star-Pathfinding repo](https://github.com/olokobayusuf/3D-A-Star-Pathfinding),
  [Unity Answers — stairs in 3D grid A\*](https://answers.unity.com/questions/1096235/using-stairs-in-3d-grid-based-pathfinding-a.html));
  topological explorers for multi-floor indoor environments combine per-floor metric maps with
  a cross-floor topological graph ([LITE](https://arxiv.org/pdf/2507.21517)).

## 10. Multi-agent map sharing and merging

- **Gossip protocols**: "decentralized communication mechanisms where each node periodically
  exchanges state information with random neighbors, gradually propagating knowledge through
  the network," enabling "scalable, fault-tolerant communication and emergent knowledge
  convergence without central control"
  ([arXiv — Revisiting Gossip Protocols](https://arxiv.org/html/2508.01531v1)).
- Gossip layered with **CRDTs** (conflict-free replicated data types) gives "eventual
  consistency of shared state across agents … without centralized control"
  ([arXiv — Gossip-Enhanced Communication Substrate](https://arxiv.org/pdf/2512.03285)).
- Decentralized multi-robot SLAM systems perform explicit **map merging**: agents identify
  overlap, merge maps, then "collaborate by sharing keyframes and map points to expand the
  shared map" ([DVM-SLAM](https://arxiv.org/pdf/2503.04126),
  [DRACo-SLAM](https://arxiv.org/pdf/2210.00867)).
- Gossip supports "gradual convergence of local world models" via probabilistic propagation
  and anti-entropy reconciliation
  ([arXiv — Revisiting Gossip Protocols](https://arxiv.org/html/2508.01531v1)).
- Provenance matters when merging reported knowledge: reliability-conditional belief updating
  defends against unreliable/poisoned reports from other agents
  ([arXiv — When Does Belief-Based Agent Memory Help?](https://arxiv.org/html/2606.22030)).

## 11. Memory-efficient representations for many private maps

- Compact bit vectors (bitsets) "can represent capability states and can be combined quickly
  with logical operations to keep latency low"; a bitset stores one boolean per bit
  ([Cleverence — Java BitSet explained](https://www.cleverence.com/articles/oracle-documentation/bitset-java-platform-se-8-4827/)).
- Per-agent visibility grids in multi-agent contest work are plain arrays "initialized with
  values denoting unobserved cells" — the baseline flat representation
  ([MAPC 2019 — Requirement Gatherers](https://arxiv.org/pdf/2006.02816)).
- Quadtrees compress "large two-dimensional binary arrays" when content is spatially coherent;
  probabilistic quadtrees do the same for occupancy maps
  ([arXiv — 2D Visibility](https://arxiv.org/pdf/2403.06494)).
- OctoMap's compression merges identical child nodes ("octree map compression method that
  keeps the 3D models compact") — the same trick in 3D
  ([Hornung et al. 2013](https://link.springer.com/article/10.1007/s10514-012-9321-0)).
- Continuous/parametric maps (GMMap) and non-uniform cell representations cut memory further
  when grid resolution exceeds what the task needs
  ([GMMap](https://arxiv.org/pdf/2306.03740),
  [non-uniform dynamic occupancy grids](https://www.researchgate.net/publication/348367869_Efficient_dynamic_occupancy_grid_mapping_using_non-uniform_cell_representation)).
- Baseline arithmetic (uncontested, from the definitions above): a 64x64 grid is 4,096 cells;
  at 1 bit/cell a visibility layer is 512 bytes; at 1 byte/cell (e.g. int8 log-odds or a state
  enum) a full layer is 4 KiB — per agent, per layer.

## Sources

1. [Yeshwanth N — Quadtree: A Spatial Data Structure for Efficient Queries (Medium)](https://medium.com/@yeshsurya/quadtree-a-spatial-data-structure-for-efficient-queries-f4f92958881d)
2. [Hanan Samet — The Quadtree and Related Hierarchical Data Structures (ACM Computing Surveys, 1984)](http://www.cs.umd.edu/~hjs/pubs/SameCSUR84-ocr.pdf)
3. [GameDev.net — Introduction to Octrees](https://www.gamedev.net/articles/programming/general-and-gameplay-programming/introduction-to-octrees-r3529/)
4. [GameDev.net forum — Using octrees for spatial representation](https://www.gamedev.net/forums/topic/666717-using-octrees-for-spatial-representation/)
5. [GameDev.net forum — Using QuadTrees in maps](https://gamedev.net/forums/topic/463418-using-quadtrees-in-maps/)
6. [ThinkAutonomous — Occupancy Grid Mapping](https://www.thinkautonomous.ai/blog/occupancy-grid-mapping/)
7. [Naren Suri — Occupancy Grid Mapping Algorithm (Medium)](https://medium.com/@SuriNaren/occupancy-grid-mapping-algorithm-e451701da0e8)
8. [Uni Freiburg — Robot Mapping: Grid Maps and Mapping With Known Poses (course slides)](http://ais.informatik.uni-freiburg.de/teaching/ss16/robotics/slides/12-occupancy-mapping.pdf)
9. [CMU 16-831 — Occupancy Mapping lecture notes](https://www.cs.cmu.edu/~16831-f12/notes/F12/16831_lecture05_vh.pdf)
10. [Agha et al. — Confidence-aware Occupancy Grid Mapping (IROS workshop)](https://karolhausman.github.io/pdf/agha17-ws-iros.pdf)
11. [USPTO 12,523,498 — Time-aware occupancy grid mapping for robots in dynamic environments](https://image-ppubs.uspto.gov/dirsearch-public/print/downloadPdf/12523498)
12. [Hornung, Wurm, Bennewitz, Stachniss, Burgard — OctoMap: An Efficient Probabilistic 3D Mapping Framework Based on Octrees (Autonomous Robots, 2013)](https://link.springer.com/article/10.1007/s10514-012-9321-0)
13. [SciSpace — OctoMap citation record](https://scispace.com/papers/octomap-an-efficient-probabilistic-3d-mapping-framework-3dp42nnofi)
14. [GMMap: Memory-Efficient Continuous Occupancy Map (arXiv)](https://arxiv.org/pdf/2306.03740)
15. [Efficient dynamic occupancy grid mapping using non-uniform cell representation (ResearchGate)](https://www.researchgate.net/publication/348367869_Efficient_dynamic_occupancy_grid_mapping_using_non-uniform_cell_representation)
16. [Memory Proxy Maps for Visual Navigation (arXiv)](https://arxiv.org/html/2411.09893)
17. [MemoNav: Selecting Informative Memories for Visual Navigation (arXiv)](https://arxiv.org/pdf/2208.09610)
18. [MemoNav: Working Memory Model for Visual Navigation (arXiv v2)](https://arxiv.org/html/2402.19161v2)
19. [SpaceVLN: Zero-Shot VLN Agent with Online Spatial Cognitive Memory (arXiv)](https://arxiv.org/pdf/2606.08992)
20. [Frontiers in Computational Neuroscience — Learning dynamic cognitive map with autonomous navigation](https://www.frontiersin.org/journals/computational-neuroscience/articles/10.3389/fncom.2024.1498160/full)
21. [Mind Meets Space: Rethinking Agentic Spatial Intelligence (arXiv)](https://arxiv.org/pdf/2509.09154)
22. [Meta-Memory: Retrieving and Integrating Semantic-Spatial Memories (arXiv)](https://arxiv.org/html/2509.20754v1)
23. [IEEE/CAA JAS — Cognitive Navigation with Topological Memory Configuration](https://www.ieee-jas.net/article/doi/10.1109/JAS.2024.124332)
24. [Machinations.io — What is Fog of War?](https://machinations.io/glossary/fog-of-war)
25. [Didac Romero — Tile-based Fog of War implementation guide (SDL/C++)](https://didacromero.github.io/Fog-of-War/)
26. [Hagelbäck & Johansson — Dealing with Fog of War in a Real Time Strategy Game Environment (ResearchGate)](https://www.researchgate.net/publication/224491323_Dealing_with_Fog_of_War_in_a_Real_Time_Strategy_Game_Environment)
27. [Investigation of the Effect of Fog of War in StarCraft Strategy Prediction (ResearchGate)](https://www.researchgate.net/publication/313279929_Investigation_of_the_Effect_of_Fog_of_War_in_the_Prediction_of_StarCraft_Strategy_Using_Machine_Learning)
28. [The Requirement Gatherers' Approach to the 2019 Multi-Agent Programming Contest (arXiv)](https://arxiv.org/pdf/2006.02816)
29. [Belief Memory: Agent Memory Under Partial Observability (arXiv)](https://arxiv.org/html/2605.05583v1)
30. [When Does Belief-Based Agent Memory Help? (arXiv)](https://arxiv.org/html/2606.22030)
31. [Zylos Research — AI Agent Memory Architectures: From Context Windows to Persistent Knowledge](https://zylos.ai/research/2026-04-05-ai-agent-memory-architectures-persistent-knowledge/)
32. [Botea, Müller, Schaeffer — Near Optimal Hierarchical Path-Finding (HPA*)](https://www.researchgate.net/publication/228785110_Near_optimal_hierarchical_path-finding_HPA)
33. [The Analysis of Efficiency Dependence of A* and HPA* (ResearchGate)](https://www.researchgate.net/publication/272551633_The_Analysis_of_Efficiency_Dependence_of_the_Shortest_Path_Finding_Algorithms_A_and_HPA)
34. [hugoscurti/hierarchical-pathfinding — HPA* implementation (GitHub)](https://github.com/hugoscurti/hierarchical-pathfinding)
35. [hit9/QuadtreePathfinding — 2D pathfinding on quadtrees (GitHub)](https://github.com/hit9/QuadtreePathfinding)
36. [Using Quadtrees for Realtime Pathfinding in Indoor Environments (ResearchGate)](https://www.researchgate.net/publication/235443232_Using_Quadtrees_for_Realtime_Pathfinding_in_Indoor_Environments)
37. [Pathfinding in hierarchical representation of large realistic virtual terrains (ScienceDirect)](https://www.sciencedirect.com/science/article/abs/pii/S0957417419305081)
38. [Amit Patel — Map Representations (theory.stanford.edu)](http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html)
39. [Frontiers in Robotics & AI — Approaches for Efficiently Detecting Frontier Cells](https://www.frontiersin.org/journals/robotics-and-ai/articles/10.3389/frobt.2021.616470/full)
40. [Awabot — Frontier-Based Exploration for Autonomous Robot](https://awabot.com/en/autonomous-exploration-method-frontiers/)
41. [Topiwala — Frontier Based Exploration for Autonomous Robot (arXiv)](https://arxiv.org/pdf/1806.03581)
42. [Gu & Xu — Frontier Based Exploration for Map Building (UNH)](https://cs.unh.edu/~tg1034/project/TianyiGu_AutonomousMapping.pdf)
43. [Topological Frontier-Based Exploration and Map-Building Using Semantic Information (PMC)](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC6832505/)
44. [Enhancing Exploration Efficiency using Uncertainty-Aware Information Prediction (arXiv)](https://arxiv.org/pdf/2412.12825)
45. [ModDB — Software Inc.: Pathfinding on multiple floors](https://www.moddb.com/games/software-inc/features/pathfinding-on-multiple-floors)
46. [Multi-Floor Zero-Shot Object Navigation Policy (arXiv)](https://arxiv.org/pdf/2409.10906)
47. [Stairway to Success: Online Floor-Aware Zero-Shot Object-Goal Navigation (arXiv)](https://arxiv.org/pdf/2505.23019)
48. [Efficient Global Navigational Planning in 3D Structures based on Point Cloud Tomography (arXiv)](https://arxiv.org/pdf/2403.07631)
49. [olokobayusuf/3D-A-Star-Pathfinding (GitHub)](https://github.com/olokobayusuf/3D-A-Star-Pathfinding)
50. [LITE: A Learning-Integrated Topological Explorer for Multi-Floor Indoor Environments (arXiv)](https://arxiv.org/pdf/2507.21517)
51. [Revisiting Gossip Protocols: A Vision for Emergent Coordination (arXiv)](https://arxiv.org/html/2508.01531v1)
52. [A Gossip-Enhanced Communication Substrate for Agentic AI (arXiv)](https://arxiv.org/pdf/2512.03285)
53. [DVM-SLAM: Decentralized Visual Monocular SLAM for Multi-Agent Systems (arXiv)](https://arxiv.org/pdf/2503.04126)
54. [DRACo-SLAM: Distributed Robust Acoustic Communication-efficient SLAM (arXiv)](https://arxiv.org/pdf/2210.00867)
55. [An Efficient Solution to the 2D Visibility Problem in Cartesian Grid Maps (arXiv)](https://arxiv.org/pdf/2403.06494)
56. [Cleverence — Java BitSet Explained](https://www.cleverence.com/articles/oracle-documentation/bitset-java-platform-se-8-4827/)
