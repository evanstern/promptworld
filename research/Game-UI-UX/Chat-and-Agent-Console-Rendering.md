---
title: Chat and Agent Console Rendering
aliases: [ChatGPT UI, Claude Code UI, Agent Loop UX]
tags: [chat-ui, agent-ux, streaming, terminal, claude-code]
type: note
created: 2026-07-25
updated: 2026-07-25
related: ["[[Game-UI-UX]]", "[[LLM-Agent-Sim-Interfaces]]", "[[TUI-and-Roguelike-UI-Craft]]", "[[Recurring-Interface-Patterns]]"]
---

# Chat and Agent Console Rendering

How conversational AI interfaces render chat turns and agent back-and-forth — ChatGPT's web
conventions, Claude Code's terminal rendering, and the documented patterns for showing an
agent "doing things." Facts cited in [[_grounding]] §Conversational AI interfaces.

## Chat-turn conventions (web)

- **Streaming is the substrate**: token-by-token rendering cuts first-token latency ~90% vs
  full-response delivery; design analyses call it "a trust mechanism" (activity signal), not
  just a performance optimization ([Medium](https://medium.com/@connect.hashblock/streaming-tokens-triple-speed-4704da5afb4d);
  [925 Studios](https://www.925studios.co/blog/chatgpt-interface-design-breakdown)).
- **Streaming has documented failure modes**: NN/g found long streamed answers with autoscroll
  "impossible… to read" and overwhelming; users want scannable, formatted, front-loaded
  answers ("truncated pyramid") ([NN/g](https://www.nngroup.com/articles/less-chat-more-answer/)).
- ChatGPT uses document-style turns (not bubbles) in a fixed-width ~65-characters-per-line
  column; affordances include per-message edit, regenerate with version selector, and stop
  generation ([925 Studios](https://www.925studios.co/blog/chatgpt-interface-design-breakdown);
  [OpenAI forum threads](https://community.openai.com/t/restore-chatgpts-stop-button-without-errors/1132068)).
- NN/g's 425-interaction study identified six conversation types requiring "varied
  interfaces," and documents "prompt controls" (UI around the input box) as the emerging
  hybrid pattern ([NN/g](https://www.nngroup.com/articles/AI-conversation-types/), [prompt controls](https://www.nngroup.com/articles/prompt-controls-genai/)).

## Claude Code's terminal rendering

- Stack: TypeScript/React on Ink ("React for CLIs") with Yoga flexbox layout — chosen because
  terminals come in all sizes, so a real layout system is required
  ([Pragmatic Engineer](https://newsletter.pragmaticengineer.com/p/how-claude-code-is-built);
  [Ink](https://github.com/vadimdemedes/ink)).
- The renderer is a heavily customized fork: screen buffers in packed typed arrays,
  damage-aware cell diffing, adjacent writes merged into minimal escape sequences, flushed
  in a single write — reported to sustain 60fps on a 200-column terminal while streaming
  ([dev.to renderer analysis](https://dev.to/vilvaathibanpb/how-claude-code-uses-react-in-the-terminal-2f3b);
  [dev.to source study](https://dev.to/minnzen/i-studied-claude-codes-leaked-source-and-built-a-terminal-ui-toolkit-from-it-4poh)).
- Documented turn-rendering behaviors: tool calls collapse by default ("Called slack 3
  times") with Ctrl+O expanding a full transcript viewer; Esc interrupts mid-turn with work
  preserved; Shift+Tab cycles permission modes including plan mode; Ctrl+T shows the agent's
  live todo checklist; the status line exposes context-window used/remaining percentages
  ([official docs](https://code.claude.com/docs/en/interactive-mode); [statusline](https://code.claude.com/docs/en/statusline)).
- Permission-prompt data point: users approve 93% of prompts — documented approval fatigue —
  motivating classifier-delegated "auto mode"
  ([Anthropic engineering](https://www.anthropic.com/engineering/claude-code-auto-mode)).
- Subagents are isolated instances that "do one focused thing, and report back a summary"
  to the main transcript ([grounding](https://joseparreogarcia.substack.com/p/claude-code-agents-explained)).

## Agent-loop patterns across tools

- **Human-in-the-loop approval**: pause execution durably, present the pending action as a
  renderable card, resume on approval — LangGraph's interrupt/checkpoint model; enables
  approve/edit/ask/time-travel ([LangChain HITL docs](https://docs.langchain.com/oss/python/langchain/frontend/human-in-the-loop);
  [Building LangGraph](https://www.langchain.com/blog/building-langgraph)).
- **Collapsible reasoning**: think-tag content rendered in a collapsible "Thinking" element
  (Open WebUI); ChatGPT shows summarized chain-of-thought for o1 while hiding raw tokens,
  criticized as an interpretability step backwards
  ([Open WebUI docs](https://docs.openwebui.com/features/chat-conversations/chat-features/reasoning-models/);
  [Simon Willison](https://simonwillison.net/2024/Sep/12/openai-o1/)).
- **Intermediate-event streaming** is documented as a production requirement distinct from
  token streaming; a known gap is tool-call rendering latency (spinner until the full JSON
  args arrive), with speculative skeleton rendering proposed
  ([vercel/ai issue](https://github.com/vercel/ai/issues/13469)).

## Terminal rendering tech facts

- Ink's `<Static>` component writes completed output permanently into native scrollback while
  only the bottom live region rerenders — the documented mechanism for accumulating a
  transcript in a terminal ([Ink](https://github.com/vadimdemedes/ink)).
- Bubble Tea (Go) uses the Elm model-update-view architecture with a cell-based renderer and
  supports inline vs full-window modes; background goroutines feed messages that trigger
  re-renders ([bubbletea](https://github.com/charmbracelet/bubbletea)).
- Textual (Python) styles TUIs with CSS-like rules and runs the same app in terminal or
  browser; its showcase includes Toad, an AI coding console ([textual.textualize.io](https://textual.textualize.io/)).

## Grounding

- [[_grounding]] — §"UI/UX of Conversational AI Interfaces" (all claims above).
