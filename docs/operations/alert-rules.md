# Alert Rule Operations

NodeScope alert rules are created and changed through role-checked server procedures. The browser may present an editable policy, but it does not determine authorization; only an Operator or Administrator equivalent can save a rule.

| Scope | Required target | Rejected target form | Intended use |
|---|---|---|---|
| `fleet` | No host identifier | Any host identifier | A threshold that applies to all eligible hosts. |
| `host` | A known NodeScope host identifier | Missing or unknown host identifier | A platform- or workload-specific threshold. |

The server rejects incoherent scope and target pairs before creating an audit record. Rule values still require an eligible metric with a production-quality evidence source at evaluation time. Experimental, stale, unavailable, or unsupported readings must not be converted into a passing alert condition or an inferred numeric value.

> Default thresholds are a starting point, not a host qualification result. Administrators must review actual Framework and Asus evidence, hardware semantics, and observed workload behavior before enabling production notifications.
