# AI Agent Collaboration Guide

This project is designed for multi-agent pair programming. To maintain consistency across Local, Server, and Remotes, follow these instructions.

## 🧱 The .agent Folder
- **rules.md**: Read this first in every session. It contains the "Ground Truth" about the project environment.
- **workflows/**: Contains predefined sequences of commands (Slash Commands) to handle repetitive tasks like synchronization or deployment.

## 🤝 Interaction Rules
1. **Consistency First**: Before making significant changes, run the `/sync` workflow to ensure you are starting from a clean, unified state.
2. **Double Remotes**: This project uses GitHub for source control and GitLab for CI/CD. **Always push to both remotes** if you are not using the automated mirror.
3. **Environment Aware**: Recognise that you are working with a local development environment (PC) and a remote production environment (Debian Server). Verify which environment you are targeting before running commands.
4. **No Destructive Resets**: Do not perform `git reset --hard` on production environments unless it is a recovery scenario approved by the user.

## 🚀 Available Workflows
- `/sync`: Aligns Local, GitHub, GitLab, and Server.
- `/deploy`: Triggers the GitLab build and deployment pipeline.
