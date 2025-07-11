<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

This repository is for Actioneer, a modular CI/CD engine written in Go.

- Use idiomatic Go, modular design, and keep extensibility in mind.
- Organize code using Go best practices: business logic in `internal/`, reusable packages in `pkg/`, and entry points in `cmd/`.
- Favor interfaces and dependency injection for extensibility.
- Write clear, concise, and well-documented code.
- Prefer composition over inheritance.
- Use standard Go error handling and logging patterns.
- When adding new features, consider how they can be extended or replaced by users.
- Use Go modules for dependency management.
- Follow the existing project structure and naming conventions.
- Write tests for new functionality and keep code testable.
- Keep the README up to date with any changes to features, usage, or setup instructions.

Actioneer aims to be a flexible, maintainable, and extensible CI/CD engine. All contributions should align with these goals.
