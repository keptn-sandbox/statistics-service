# Release Notes 0.2.0

* Fixed memory leak [#4](https://github.com/keptn-sandbox/statistics-service/issues/9)
* Changed API responses to a more user friendly format [#5](https://github.com/keptn-sandbox/statistics-service/issues/8)
* Set distributor image to version 0.7.3

Example payload with new response format:
```json
{
  "from": "2020-09-21T02:41:45Z",
  "to": "2020-09-23T14:02:42Z",
  "projects": [
    {
      "name": "sockshop",
      "services": [
        {
          "name": "carts",
          "events": [
            {
              "type": "sh.keptn.event.approval.finished",
              "count": 1
            },
            {
              "type": "sh.keptn.event.approval.triggered",
              "count": 1
            }
          ],
          "keptnServiceExecutions": [
            {
              "name": "gatekeeper-service",
              "executions": [
                {
                  "type": "sh.keptn.event.approval.finished",
                  "count": 1
                },
                {
                  "type": "sh.keptn.event.approval.triggered",
                  "count": 1
                }
              ]
            },
            {
              "name": "helm-service",
              "executions": [
                {
                  "type": "sh.keptn.event.deployment.finished",
                  "count": 1
                }
              ]
            }
          ]
        }
     ] 
    }
  ]
}
```
