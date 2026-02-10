# 📖 Agentic OS: User Manual

This manual explains how to leverage your AI agent's native capabilities to build, deploy, and maintain the **Vera Massage Bot**.

---

## 🏗️ 1. The Multi-Layer Context System

We use three native layers to manage how I think and act:

### 📏 Rules (`.agent/rules/`)

* **What they are**: Passive "Subconscious" constraints.
* **How to use**: No action needed. I automatically follow these in every interaction.
* **Key Rules**:
  * **Logic over Compliance**: I will challenge "bad" ideas and propose better engineering solutions.
  * **Hypothesis First**: I will never change code without explaining my "Why" first.

### 🛠️ Workflows (`.agent/workflows/`)

* **What they are**: Active "Macro" commands you trigger via slash-commands.
* **How to use**: Type `/command` in the chat.
* **Key Workflows**:
  * `/checkpoint`: Consolidates progress, updates changelogs, and creates handoffs.
  * `/report`: Pulls live metrics from the bot and formats them into a BI summary.

### 🧠 Skills (`.agent/skills/`)

* **What they are**: Specialized "Expertise" I "equip" based on your intent.
* **How to use**: Just mention a topic (e.g., "Check the database" or "Update the UI").
* **Key Skills**:
  * **TWA Aesthetics**: Activates for premium CSS/JS UI work.
  * **Database Expert**: Activates for safe Postgres migrations.
  * **DevOps Harness**: Activates for deployment and testing tasks.

---

## 🗣️ 2. Communication Protocols

### "Logic Over Compliance"

If you suggest a change that might break common patterns or introduce debt, I will:

1. **Acknowledge** the request.
2. **Explain** the risks (Security, Scalability, Consistency).
3. **Propose** a "Senior Engineer" alternative.
4. **Wait** for your final decision.

### "State Your Hypothesis"

Before I modify any file, I will follow this loop:

1. **Observe**: "I see symptoms X and Y."
2. **Hypothesize**: "I suspect the root cause is Z."
3. **Propose**: "I will test this by doing A."
4. **Execute**: Only after you say "Go."

---

## 🔄 3. Maintenance Rituals

### End-of-Session (`/checkpoint`)

Always run `/checkpoint` before closing the tab.
* It ensures the next session has a clear starting point.
* It keeps the `Project-Hub.md` and `CHANGELOG.md` updated.

### Health Checks (`/report`)

Run `/report` regularly to see if the bot is actually being used and if there are any technical errors in production.

---

## 🛠️ 4. Expanding the OS

You can add new capabilities yourself:
* **New Rule?** Create a `.md` file in `.agent/rules/`.
* **New Workflow?** Create a `.md` file in `.agent/workflows/`.
* **New Skill?** Create a directory in `.agent/skills/` with a `SKILL.md`.

> [!TIP]
> This system is designed to make me **proactive**. If you feel I'm being too "obedient" and not "engineering" enough, remind me of the **Logic Over Compliance** rule!
