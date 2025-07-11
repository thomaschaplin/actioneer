<img src="assets/logo.jpeg" alt="logo" width="256" height="256" />

# Actioneer

Actioneer is a lightweight, blazing-fast CI/CD engine written in Go. It enables you to define and run pipelines using simple YAML workflows, supporting various built-in actions commonly found in pipeline automation.

## Features
- **Modular architecture:** Easily add or extend actions.
- **Configurable pipelines:** Define workflows in YAML.
- **Built-in actions:**
  - `actions/checkout@v1`: Clone a git repository.
  - `actions/cat@v1`: Output file contents.
  - `actions/shell@v1`: Run shell commands.
  - `actions/env@v1`: Set environment variables.
  - `actions/env-secure@v1`: Set secure environment variables.
  - `actions/env-unset@v1`: Unset environment variables.
  - `actions/print-env@v1`: Print environment variables.
  - `actions/upload-artifact@v1`: Upload files as artifacts.
  - `actions/download-artifact@v1`: Download artifacts.
- **Extensible:** Add your own actions in Go with minimal effort.

## Architecture
- **cmd/**: Main entry point (`main.go`).
- **internal/**: Core logic, actions, and config loading.
- **pkg/**: (Reserved for public packages/extensions.)
- **.actioneer/workflows/**: Example workflow YAMLs.

## Getting Started

1. **Install Go:** https://golang.org/doc/install
2. **Clone the repo:**
   ```sh
   git clone https://github.com/thomaschaplin/actioneer.git
   cd actioneer
   ```
3. **Run a workflow:**
   ```sh
   go run ./cmd/main.go <workflow.yaml>
   # Example:
   go run ./cmd/main.go checkout.yaml
   ```
   Workflows are in `.actioneer/workflows/`.

## Example Workflows

### 1. Checkout and Cat Files
```yaml
pipeline:
  name: Checkout and Cat Files Example
  steps:
    - name: checkout
      uses: actions/checkout@v1
      url: https://github.com/thomaschaplin/kilokeeper.git
    - name: cat files
      uses: actions/cat@v1
      files:
        - README.md
```

### 2. Shell Command
```yaml
pipeline:
  name: Shell Command Example
  steps:
    - name: checkout
      uses: actions/checkout@v1
      url: https://github.com/thomaschaplin/kilokeeper.git
    - name: run shell command
      uses: actions/shell@v1
      command: ls -l
```

### 3. Environment Variables
```yaml
pipeline:
  name: Environment Variables Example
  steps:
    - name: set basic environment variables
      uses: actions/env@v1
      env:
        FOO: bar
        HELLO: world
    - name: print environment variables
      uses: actions/print-env@v1
    - name: print environment variables with shell command
      uses: actions/shell@v1
      command: echo $FOO $HELLO
    - name: set secure environment variables
      uses: actions/env-secure@v1
      env:
        SECRET_TOKEN: supersecret
        TOKEN: abc123
    - name: print env
      uses: actions/print-env@v1
    - name: unset environment variables
      uses: actions/env-unset@v1
      files:
        - FOO
    - name: print environment variables
      uses: actions/print-env@v1
    - name: unset secure environment variables (secure)
      uses: actions/env-unset@v1
      files:
        - SECRET_TOKEN
    - name: print environment variables
      uses: actions/print-env@v1
    - name: echo environment variable (secure)
      uses: actions/shell@v1
      command: echo $TOKEN
```

### 4. Upload and Download Artifact
```yaml
pipeline:
  name: Upload and Download Artifact Example
  steps:
    - name: checkout
      uses: actions/checkout@v1
      url: https://github.com/thomaschaplin/kilokeeper.git
    - name: cat readme
      uses: actions/cat@v1
      files:
        - main.go
        - README.md
    - name: upload artifact
      uses: actions/upload-artifact@v1
      files:
        - main.go
        - README.md
    - name: checkout
      uses: actions/checkout@v1
      url: https://github.com/thomaschaplin/actioneer.git
    - name: download artifact
      uses: actions/download-artifact@v1
      files:
        - main.go
    - name: run shell command
      uses: actions/cat@v1
      files:
        - main.go
        - README.md
```

## Configuration Reference

Each workflow YAML defines a `pipeline` with a name and a list of steps. Each step can use a built-in or custom action, and may specify files, commands, environment variables, or URLs as needed.

## Performance and Speed

Actioneer is designed for speed and efficiency. It executes pipelines quickly by running each step sequentially with minimal overhead, leveraging Go's fast execution and low memory footprint. Built-in actions are optimized for common CI/CD tasks, and the engine is suitable for both local and server environments.

### Time Logging

Actioneer provides detailed timing information for each step and the overall pipeline. For every step, the engine logs the duration it took to execute, and at the end, it logs the total pipeline duration. This helps you:
- Identify slow steps in your workflow
- Optimize pipeline performance
- Monitor and compare execution times across runs

**Example log output:**
```json
{"level":"info","message":"Completed step","step":"checkout","detail":{"duration":"1.234s"}}
{"level":"info","message":"Pipeline completed","step":"Checkout and Cat Files Example","detail":{"duration":"2.345s"}}
```

All logs are in structured JSON for easy parsing and integration with log management tools.

## Supported Actions and Extensibility Philosophy

Actioneer intentionally supports only a curated set of built-in actions. This design choice ensures:
- **Efficiency:** All built-in actions are optimized for speed and minimal resource usage.
- **Reliability:** Each action is tested and maintained as part of the core engine, reducing the risk of unexpected failures.
- **Supportability:** The Actioneer team can provide help and updates for all supported actions, ensuring a consistent experience.

If your workflow requires functionality not covered by the built-in actions, you can use the `actions/shell@v1` action to run any shell command. This provides maximum flexibility for custom needs. However, for best performance and reliability, we recommend requesting a new built-in action instead of relying on shell commands. Built-in actions are more robust, portable, and easier to maintain.

To request a new action, please open an issue or pull request on the Actioneer repository.

## Extending Actioneer
- Add new actions by implementing an `ActionFunc` in Go and registering it in `internal/actions.go`.
- See the built-in actions for examples.

## Running Actioneer

Actioneer can be run in two modes:

- **CLI:** Run workflows directly from the command line.
- **Webhook server:** Run as a server and trigger workflows via HTTP POST requests.

### CLI Usage

Run a workflow YAML directly:

```sh
go run ./cmd/main.go <workflow.yaml>
# Example:
go run ./cmd/main.go shell.yaml
```

### Webhook Usage

Start Actioneer in webhook mode (default port: 8080):

```sh
go run ./cmd/main.go --webhook
```

You can then trigger a workflow by sending a POST request to `/webhook`:

```sh
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"workflow":"<workflow.yaml>"}'
# Example:
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{"workflow":"shell.yaml"}'
```

This will execute the specified workflow (e.g., `shell.yaml`) from your `.actioneer/workflows/` directory.

## Testing

Actioneer includes automated tests for core actions and configuration loading to ensure reliability and maintainability.

### Running Tests

To run all tests:

```sh
go test ./internal/... -v
```

This will execute all unit tests in the `internal/` directory and display verbose output.

### Writing Tests

- Add new tests in files ending with `_test.go` alongside the code being tested (e.g., `internal/actions_test.go`).
- Use Go's standard `testing` package.
- Cover new features and edge cases to help maintain code quality.

### Test Philosophy

Tests are required for all new features and bug fixes. Comprehensive test coverage ensures Actioneer remains robust and extensible as it evolves.

## License
MIT
