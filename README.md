Feature Flag Gatekeeper is a service that makes controllably limited feature access decisions. Enforcement of those decisions remains the responsibility of consumers.

## Problems & Project Solutions
Problem: Features are integrated abruptly either to all the users or none, which leads to destructive results in case if the feature is bad.
Solution: The destructions are localized, because only the defined share of the users get an access to a new feature.

Problem: After feature full integration an external factor may affect the statistics of the change, which leads to incorrect assumptions and decisions.
Solution: The versions with and without a feature can exist simultaneously, which allows to track the control groups with clear differences in feature set.

Problem: Abrupt feature integration creates massive resistance to change, which creates additional social and reputational costs from syncronized backslash.
Solution: Feature exposure can be controlled dynamically, which enables gradual mass adaptation and prevents sharp global irritation.
