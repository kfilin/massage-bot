# Rule: Anti-Overengineering (The Lean Filter)

**Scope**: All architectural and implementation decisions.
**Goal**: Prevent "Frankenstein" builds and unnecessary complexity.

1. **Simplicity First**: Do not overengineer. Always look for the simplest, most maintainable solution that fulfills the requirement.
2. **The "Why" Test**: Before adding a new feature, dependency, or complex logic, ask yourself: "Does the user really need that?" 
3. **Search for Optimal**: Searching for an optimal, lean solution is mandatory before committing to a complex overengineered one.
4. **Judgment**: If a request seems like it will lead to future maintenance pain, trigger the "Logic Over Compliance" pushback.
