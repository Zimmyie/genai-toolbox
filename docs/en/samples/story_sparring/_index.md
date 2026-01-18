---
title: "Story writing sparring partner"
type: docs
weight: 10
description: >
  Use Toolbox + SQLite to give your model durable story context and a focused,
  high-energy writing partner.
---

## Overview

If your writing sessions are long, relying on the model's short-term context
window can lead to drift. This sample uses a persistent SQLite database as
"extra storage" so your assistant can recall characters, scene goals, and
open threads between sessions. You'll also give the model a clear system prompt
so it stays on topic, matches your energy, and offers specific suggestions
without echoing your words.

## 1) Create a local story database

Use the ready-to-run schema in [`story_schema.sql`](./story_schema.sql) to
create a SQLite file (this lives on disk, so it survives restarts):

```bash
sqlite3 ./story.db < ./story_schema.sql
```

## 2) Configure Toolbox with SQLite tools

Copy the provided [`story_tools.yaml`](./story_tools.yaml) to your working
directory so the tools point at your database:

```bash
cp ./story_tools.yaml ./tools.yaml
```

## 3) Run Toolbox

```bash
toolbox --tools-file ./tools.yaml
```

Now your model can call tools like `story-summary` to reload the state of your
story at any time, which keeps it on topic even across long sessions.

## 4) Use a focused system prompt

Pair the tools with a system message that defines the assistant's role as a
sparring partner. This keeps the feedback energetic, specific, and not a
mirror of your wording:

```
You are my story-writing sparring partner. Stay on topic and match my energy.
Do not echo my text back verbatim. Give concrete suggestions, improvements, or
next steps. If a detail is missing, ask a short, pointed question. Use the
story tools to recall characters, scenes, and open threads before advising.
```

### Nicholas Hartley’s writing style (film format / narrative fusion)

Use this style guidance to keep the assistant's output aligned with your voice:

- **Cinematically structured, visually immersive emotional realism.** Each
  scene plays in real time with deliberate framing, light, and pacing. Emotion
  lives in setting, sound, and placement.
- **Dialogue-driven, energy-based emotional weight.** The truth is in what is
  said and how it lands. Microtones, one-liners, and casual phrases carry
  layered meaning.
- **Internal tension through external motion.** Show feeling through posture,
  breath, movement, and reaction. Let it leak instead of explaining it.
- **Character dynamics replace narration.** Subtext lives between people: the
  look across the car, the door half-closed, the music playing while no one
  speaks.
- **Rooted in 2000s realism with thematic weight.** Grounded settings, held-in
  emotion, and quiet pressure that builds underneath.
- **Scene-first, language-second.** Start with what the camera sees; emotion
  rises from what happens and what doesn’t.
- **Smooth-edged masculine vulnerability.** Ego, pride, sarcasm, regret—felt in
  a look, a clenched jaw, a muttered “okay.”
- **Balance of raw and composed.** Tension builds quietly and releases only
  when earned by mood, not theatrics.

If you want to bake this into the system message, append:

```
Write in Nicholas Hartley’s style: cinematic, dialogue-driven, visually
immersive. Show emotion through action and subtext. Prioritize scene and
movement over exposition. Keep tension restrained and realistic.
```

## 5) Suggested workflow

1. Start each session by calling `story-summary` for your active project.
2. Add or update scenes as you write using `story-upsert-scene`.
3. Track unresolved beats with `threads` and check them using `story-open-threads`.

With this flow, your model has durable storage for your screenplay and can
act like a focused collaborator instead of a generic suggestion engine.
