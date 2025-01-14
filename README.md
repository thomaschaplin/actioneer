<img src="assets/logo.jpeg" alt="logo" width="256" height="256" />

# Actioneer

Actioneer is a lightweight, blazing-fast GitHub Actions clone built in Go. It provides a seamless way to define, orchestrate, and run CI/CD workflows locally or in your self-hosted environment. With Actioneer, developers can manage automated tasks and pipelines without relying on external services, ensuring greater flexibility, speed, and control over their DevOps processes.

**Key Features:**

- YAML-based workflow definitions, just like GitHub Actions
- Native Go performance and concurrency
- Modular design for easy customization and extension
- Ideal for self-hosted CI/CD solutions or local workflow testing

Whether you're building a personal project or moving away from hosted services such as GitHub Actions or Jenkins, Actioneer has you covered! 🚀

## Setup

1. `go run main.go` to run the program
2. `smee -u https://smee.io/1eOP0ZmjbtoBY9XS -p 8080` to run the smee client to forward the webhook to the local server
3. Visit `https://smee.io/1eOP0ZmjbtoBY9XS` to see the webhook events
4. Trigger the webhook by pushing a commit to the repository which has the webhook configured
5. The webhook event should be visible in the smee client and the local server should log the event
6. If you need to rerun the POST you can do so via the smee web interface

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
