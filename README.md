<img src="assets/logo.jpeg" alt="logo" width="256" height="256" />

# Actioneer

Actioneer is a lightweight, blazing-fast GitHub Actions clone built in Go. It provides a seamless way to define, orchestrate, and run CI/CD workflows locally or in your self-hosted environment. With Actioneer, developers can manage automated tasks and pipelines without relying on external services, ensuring greater flexibility, speed, and control over their DevOps processes.

**Key Features:**

- YAML-based workflow definitions, just like GitHub Actions
- Native Go performance and concurrency
- Modular design for easy customization and extension
- Ideal for self-hosted CI/CD solutions or local workflow testing

Whether you're building a personal project or moving away from hosted services such as GitHub Actions or Jenkins, Actioneer has you covered! 🚀

## Development Setup

Install the following dependencies:
- [Go](https://golang.org/doc/install)
- [smee-client](https://github.com/probot/smee-client)

1. Clone the repository
2. Run `export ACTIONEER_WEBHOOK_SECRET="your_secret"` to set the webhook secret
3. Run `go run main.go` to start the server
4. Visit `https://smee.io` to start a new webhook and copy the URL and ID
5. Create a [new webhook](https://docs.github.com/en/webhooks/using-webhooks/creating-webhooks) in a GitHub repository and paste the URL from step 2
6. Keep the website open to view the webhook events as they come in
7. Run `smee -u https://smee.io/<ID> -p 8080` to forward the webhook to the local server
8. `smee -u https://smee.io/<ID> -p 8080` to run the smee client to forward the webhook to the local server
9. Trigger the webhook by pushing a commit to the repository which has the webhook configured
10. The webhook event should be visible in the smee client and the local server should log the event
11. If you need to rerun the POST you can do so via the smee web interface

## Usage

Actioneer uses a YAML-based syntax to define workflows. Here's an example of a simple workflow that runs some bash commands:

```yaml
jobs:
  # Use the below "action" to checkout another repository
  # checkout:
  #   steps:
  #     - action: checkout
  #       repo: "https://github.com/thomaschaplin/actioneer.git"
  list_files:
    steps:
      - action: run
        command: ls -lrta
      - action: run
        command: ls actions
      - action: run
        command: cat .gitignore
  cat_readme:
    steps:
      - action: run
        command: cat README.md
```

## Roadmap

- [ ] Workflow Schema
  - [x] Jobs
  - [x] Steps
  - [x] Actions
  - [ ] Events
- [ ] Actions
  - [x] Run (bash)
  - [x] Checkout (git clone)
  - [ ] AWS Commands
  - [ ] Docker Commands
- [ ] Workflow Events
  - [x] Push
  - [ ] Pull Request
  - [ ] Issue
  - [ ] Release
- [ ] Kubernetes Integration
  - [ ] Webhook service to spin up a new pod for each webhook event
- [ ] Examples
  - [ ] Simple CI/CD Pipeline
- [ ] Split types for each action (example: checkout doesn't need "command" field)
