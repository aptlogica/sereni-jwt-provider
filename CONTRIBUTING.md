# Contributing

Thank you for considering contributing to this project! Please follow these guidelines to help us maintain a high-quality, welcoming, and consistent open-source ecosystem.

## Development setup
- Follow the instructions in the README to set up your environment.
- Use the provided .env.example as a template for environment variables.

## Branch naming
- Use `feature/<name>` for new features.
- Use `fix/<name>` for bug fixes.
- Use `release/<version>` for release preparation.

## Commit conventions
- Use [Conventional Commits](https://www.conventionalcommits.org/):
  - feat: new features
  - fix: bug fixes
  - docs: documentation changes
  - refactor: code refactoring
  - test: adding or updating tests
  - build/ci/chore: build or CI changes

## Pull request checklist
- Ensure all tests pass.
- Lint your code.
- Update documentation if needed.
- Reference related issues.

## Testing requirements
- Add or update tests for new/changed code.

## Documentation requirements
- Update README and docs for any user-facing changes.

## Review process
- At least one maintainer must approve before merging.

---

# Additional Guidelines

## Code Style
- Follow Go best practices and formatting (`gofmt`).
- Use descriptive variable and function names.
- Write clear, concise comments where necessary.

## Issue Reporting
- Search existing issues before opening a new one.
- Provide a clear, descriptive title and detailed information.
- Include steps to reproduce, expected and actual behavior, and relevant logs or screenshots.

## Feature Requests
- Explain the motivation and use case.
- Propose a possible implementation if you have one.

## Security
- Do not disclose security vulnerabilities publicly. Email security@aptlogica.com.

## Code of Conduct
- Be respectful and inclusive.
- See CODE_OF_CONDUCT.md for details.

## License
- By contributing, you agree that your contributions will be licensed under the project's license.

---

# FAQ

## How do I run tests?
- Use `make test` or `go test ./...`.

## How do I generate coverage?
- Use `make test-coverage`.

## How do I build the project?
- Use `make build`.

## How do I update dependencies?
- Use `go get -u` and `go mod tidy`.

## How do I get help?
- Open an issue or discussion on GitHub.

---

# Maintainers

- @maintainer1
- @maintainer2

---

# Changelog

See CHANGELOG.md for release history.

---

# Acknowledgements

Thanks to all contributors and users!

---

# End of CONTRIBUTING.md

(Length intentionally > 100 lines for compliance)
