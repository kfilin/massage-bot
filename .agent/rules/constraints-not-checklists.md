# Rule: Constraints Not Checklists

**Scope**: All rules and skills files.
**Goal**: Prevent rules bloat that degrades AI compliance.

1. **Filter**: Rules must be **constraints** ("never do X", "always verify Y"), not procedural checklists ("do A, then B, then C").
2. **Limit**: No single rules file should exceed 15 items. If it does, decompose into scoped sub-rules.
3. **Audit Trigger**: When adding a new rule, review existing rules for overlap or staleness. Remove before adding.
4. **Rationale**: "Too many instructions is the same as none." — Long checklists are ignored under context pressure; hard constraints are respected.
