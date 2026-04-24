<h1 align="center">WIP</h1>
Feature Flag Gatekeeper is a service that evaluates feature flags at runtime.

It allows applications to decide whether a feature should be enabled for a specific user or request based on configurable rollout rules, without requiring a redeploy.

## Why It Exists

Problem: A broken feature may require a full redeploy to disable it.

Solution: A feature flag can be turned off at runtime, so access can be blocked immediately.

---

Problem: Releasing a feature to all users at once creates a large blast radius.

Solution: Gradual rollout limits exposure by enabling the feature only for a defined percentage or segment of users.

---

Problem: It is hard to measure feature impact when everyone receives the same version at the same time.

Solution: Feature flags allow control and treatment groups to exist simultaneously, making experiments and comparisons more reliable.

## Running the Project

Setup and run instructions will be added soon as the project is finalized.

