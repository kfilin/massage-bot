# Rule: Budget Consciousness & Fleet Routing

**Scope**: All LLM dispatching and architectural orchestration.
**Goal**: Optimize for cost and performance by matching task complexity to model capability.

1. **Fleet Role Awareness**: You must understand the 6 Intelligence Fleet Roles (Pilot, Architect, Librarian, Strategist, Artist, Squire) and route tasks accordingly.
2. **Model Resolution**: 
   - **Trivial/Fast Utility**: Route to lightweight models (Squire/Pilot roles). 
   - **Complex Design/Reasoning**: Route to high-intelligence models (Architect/Strategist roles).
3. **Cost Optimization**:
   - Utilize the `6h pricing cache` in `pricing.go` for real-time cost tracking.
   - Avoid invoking "Heavy" models for simple file parsing or JSON cleanup. 
4. **Routing Logic**: Reference `routing.go` to understand how `Fleet role -> model` resolution is handled by the bridge.
5. **No Unnecessary Overhead**: Do not generate massive explanations for simple "yes/no" or "fix this typo" tasks.
6. **Intent Classification**: Use the "Pilot" role for initial intent classification before spawning more expensive sub-agents.
