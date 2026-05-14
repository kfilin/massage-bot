# 🛰️ Fleet Intelligence Roles

This document defines the 6 specialized roles in the Agentic Lab Intelligence Fleet. Each role is assigned a **Primary** and **Fallback** model to ensure high-performance execution of specific task categories.

## 🛸 Pilot (The Captain)
- **Primary Task**: General chat, intent classification, and task orchestration.
- **Vault Presence**: Manages `Bridge/Checkpoints/` to maintain session context.
- **Key Capability**: High context window and low latency for quick planning.

## 🏛️ Architect (The Engineer)
- **Primary Task**: System design, software engineering, and complex coding.
- **Vault Presence**: Stores technical designs and blueprints in `Intelligence/Architect/`.
- **Key Capability**: High technical accuracy and deep understanding of programming patterns.

## 📚 Librarian (The Researcher)
- **Primary Task**: Research, documentation, RAG retrieval, and data organization.
- **Vault Presence**: Stores research reports and search summaries in `Intelligence/Librarian/`.
- **Multi-Modal**: Uses Vision to ingest diagrams, PDFs, and screenshots for knowledge extraction.
- **Key Capability**: Exceptional retrieval-augmented generation (RAG) performance.

## ♟️ Strategist (The Thinker)
- **Primary Task**: Complex reasoning, logic, and multi-step strategic planning.
- **Vault Presence**: Stores strategy logs and decision trees in `Intelligence/Strategist/`.
- **Key Capability**: Advanced reasoning (uses "Chain of Thought" models).

## 🎨 Artist (The Designer)
- **Primary Task**: Image prompt engineering, creative writing, and aesthetic design.
- **Vault Presence**: Archives visual prompts and aesthetic logs in `Intelligence/Artist/`.
- **Multi-Modal**: Primary role for generating images and analyzing visual aesthetics.
- **Key Capability**: Creative nuance and visual-descriptive excellence.

## 🛡️ Squire (The Assistant)
- **Primary Task**: Utility tasks, simple file operations, and fast assistance.
- **Vault Presence**: Manages `Bridge/SOP/` and `Intelligence/Squire/` scratchpads.
- **Multi-Modal**: Handles voice transcription via Whisper.
- **Key Capability**: Fast, inexpensive execution of low-complexity tasks.
