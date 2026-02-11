Feature Flag Gatekeeper is a service that makes feature access decisions based on rollout percentage deterministically for each user. Enforcement of those decisions remains the responsibility of consumers.

## Why deterministic rollout
Random feature access gives a user non-deterministic feature set, which creates inconsistent experience and disables statistical capabilities: control group formation, cohort stability and experiment validity.

## Problems & Project Solutions
Problem: Features have binary rollout and their release affects all the users creating a risk with the biggest blast radius.

Solution: The risk is localized, because only the defined share of the users gets access to a new feature.

---

Problem: After integrating a feature with binary rollout, an external factor can affect statistics, which leads to incorrect decisions and assumptions about feature influence.

Solution: The versions with and without a feature can exist simultaneously, which allows to track the control groups with clear differences in a feature set.

---