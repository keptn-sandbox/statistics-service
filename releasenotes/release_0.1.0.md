# Release Notes 0.1.0

Initial release for the statistics service, which collects statistics of Cloud-Events and Keptn-service usage.

Example:
```json
{
  "from": "2020-09-21T02:41:45Z",
  "to": "2020-09-23T14:02:42Z",
  "projects": {
    "ck-sockshop": {
      "name": "sockshop",
      "services": {
        "carts": {
          "name": "carts",
          "executedSequences": 1,
          "events": {
            "sh.keptn.event.approval.finished": 1,
            "sh.keptn.event.approval.triggered": 1,
            "sh.keptn.event.configuration.change": 3,
            "sh.keptn.events.deployment-finished": 1,
            "sh.keptn.events.evaluation-done": 1,
            "sh.keptn.events.tests-finished": 1
          },
          "keptnServiceExecutions": {
            "gatekeeper-service": {
              "name": "gatekeeper-service",
              "executions": {
                "sh.keptn.event.approval.finished": 1,
                "sh.keptn.event.approval.triggered": 1,
                "sh.keptn.event.configuration.change": 2
              }
            },
            "helm-service": {
              "name": "helm-service",
              "executions": {
                "sh.keptn.events.deployment-finished": 1
              }
            },
            "https://github.com/keptn/keptn/cli#configuration-change": {
              "name": "https://github.com/keptn/keptn/cli#configuration-change",
              "executions": {
                "sh.keptn.event.configuration.change": 1
              }
            },
            "jmeter-service": {
              "name": "jmeter-service",
              "executions": {
                "sh.keptn.events.tests-finished": 1
              }
            },
            "lighthouse-service": {
              "name": "lighthouse-service",
              "executions": {
                "sh.keptn.events.evaluation-done": 1
              }
            }
          }
        }
      }
    }
  }
}
```
