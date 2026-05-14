---
name: multi_modal_intelligence
description: Protocol for handling non-textual inputs (Images, Voice, Diagrams) within the Intelligence Fleet.
category: Orchestration
---

# Skill: Multi-Modal Intelligence Protocol

## 1. Vision Analysis (Images & Diagrams)
When the user sends an image, the **Pilot** must route it to the **Artist** (for aesthetic/creative analysis) or the **Librarian** (for data/text extraction).

### 🛠️ Execution Steps:
1. **Model Selection**: Ensure a vision-capable model is active (e.g., `google/gemini-2.0-flash-001` or `openai/gpt-4o`).
2. **Analysis**: Describe the image in detail. If it contains text, perform OCR and format it as Markdown.
3. **Vault Logging**: Every analysis MUST be logged in `Intelligence/[Role]/Visual-Logs/`.
   - File Name: `IMG-YYYY-MM-DD-HHMM.md`
   - Content: Analysis summary, extracted text, and a reference to the original file ID.

## 2. Voice Intelligence (Whisper)
Voice memos are processed by the **Squire** to reduce overhead for the reasoning models.

### 🛠️ Execution Steps:
1. **Transcription**: Use the `whisper` container via the bridge API.
2. **Actionability**: Convert the transcript into a task, note, or command.
3. **Vault Logging**: Store raw transcripts in `Intelligence/Squire/Transcripts/` if they are longer than 30 seconds.

## 3. Multi-Modal Reasoning
Combine visual and textual context to solve complex problems (e.g., "Fix this CSS based on the screenshot").

### 🛠️ Guidelines:
- **Spatial Reasoning**: Identify the location of elements in diagrams using relative coordinates (top-left, center, etc.).
- **Aesthetic Consistency**: When analyzing UI/UX, the **Artist** must check against the project's design tokens.
