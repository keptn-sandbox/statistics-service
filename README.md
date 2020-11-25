# Statistics Service

This service provides usage statistics about a Keptn installation.

## Compatibilty Matrix

| Keptn Version    | [Statistics Service](https://hub.docker.com/r/keptnsandbox/statistics-service/tags?page=1&ordering=last_updated) | Kubernetes Versions                      |
|:----------------:|:----------------------------------------:|:----------------------------------------:|
|       0.7.1      | keptnsandbox/statistics-service:0.1.0    | 1.14 - 1.19                              |
|       0.7.2      | keptnsandbox/statistics-service:0.1.1    | 1.14 - 1.19                              |
|       0.7.3      | keptnsandbox/statistics-service:0.2.0    | 1.14 - 1.19                              |


## Deploy in your Kubernetes cluster

To deploy the current version of the *shipyard-controller* in your Keptn Kubernetes cluster, use the files `deploy/pvc.yaml` and `deploy/service.yaml` from this repository and apply it.

```console
kubectl apply -f deploy/service.yaml -n keptn
```

## Delete in your Kubernetes cluster

To delete a deployed *statistics-service*, use `deploy/service.yaml` from this repository and delete the Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml -n keptn
```

### Generate  Swagger doc from source

First, the following go modules have to be installed:

```
go get -u github.com/swaggo/swag/cmd/swag
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

If the `swagger.yaml` should be updated with new endpoints or models, generate the new source by executing:

```console
swag init
```

## How to use the service

Once the service is deployed in your cluster, you can access it using `port-forward`:

```
kubectl port-forward -n keptn svc/statistics-service 8080
``` 

You can then browse the API docs at by opening the Swagger docs in your [browser](http://localhost:8080/swagger-ui/index.html).

To retrieve usage statistics for a certain time frame, you need to provide the [Unix timestamps](https://www.epochconverter.com/) for the start and end of the time frame.
E.g.:

```
http://localhost:8080/v1/statistics?from=1600656105&to=1600696105
```

cURL Example:

```
curl -X GET "http://localhost:8080/v1/statistics?from=1600656105&to=1600696105" -H "accept: application/json"
```

*Note*: You can generate timestamps using [epochconverter.com](https://www.epochconverter.com/).

### Configuring the service

By default, the service aggregates data with a granularity of 30 minutes. Whenever this period has passed, the service will create
a new entry in the Keptn-MongoDB within the Keptn cluster. If you would like to change how often statistics are stored, you can set the 
variable `AGGREGATION_INTERVAL_SECONDS` to your desired value.

## Using the CLI

The `keptn-usage-stats` CLI allows to aggregate a set of files containing response payloads from the statistics-service and present those in a user friendly manner.

The CLI knows the following flags:
- `--folder (-f):` The folder containing the JSON files exported from the statistics-service.
- `--period (-p):` The period under consideration, one option of: [separated, aggregated]
   - `separated` - Keeps the files separated and provides the summary for each file.
   - `aggregated` - Aggregates all files to a single summary.
- `--granularity (-g):` The level of details, list of [overall, project, service], default is `overall`
- `--includeEvents:` List of events that define an automation unit, default is `all`
- `--includeServices:` List of services that are considered for an automation unit, default is `all`
- `--excludeProjects:` List of project names that are excluded from the summary
- `--export:` The format to export the statistics, supported are [json, csv]
- `--separator:` The separator used for the CSV exporter [, or ;]

### Examples

The following command will create a CSV file for each of the statistics files located in the directory `./example_payloads`:

```
$ keptn-usage-stats --folder=./example_payloads --granularity=service --export=csv --period=separated

---
Timeframe: 2020-09-21 02:41:45 +0000 UTC - 2020-09-23 14:02:42 +0000 UTC
---

Overall: stats1.json > Keptn
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- helm-service:                  1       sh.keptn.event.deployment.finished

Project: Keptn > sockshop
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- helm-service:                  1       sh.keptn.event.deployment.finished

Service: Keptn > sockshop > carts
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- helm-service:                  1       sh.keptn.event.deployment.finished

---
Timeframe: 2020-06-21 02:41:45 +0000 UTC - 2020-06-23 14:02:42 +0000 UTC
---

Overall: stats2.json > Keptn
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- argo-service:                  1       sh.keptn.event.deployment.finished

Project: Keptn > my-project
- Executions: 3
-------------------------------------------------
- argo-service:                  1       sh.keptn.event.deployment.finished
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered

Service: Keptn > my-project > users
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- argo-service:                  1       sh.keptn.event.deployment.finished
```

The resulting CSV files will look as follows:

**stats_1.csv**
```
Timeframe,Overall: stats1.json > Keptn,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished),Project: Keptn > sockshop,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished),Service: Keptn > sockshop > carts,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished)
2020-09-21 02:41:45 +0000 UTC - 2020-09-23 14:02:42 +0000 UTC,3,1,1,1,3,1,1,1,3,1,1,1
```

**stats_2.csv**
```
Timeframe,Overall: stats2.json > Keptn,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),argo-service (sh.keptn.event.deployment.finished),Project: Keptn > my-project,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),argo-service (sh.keptn.event.deployment.finished),Service: Keptn > my-project > users,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),argo-service (sh.keptn.event.deployment.finished)
2020-06-21 02:41:45 +0000 UTC - 2020-06-23 14:02:42 +0000 UTC,3,1,1,1,3,1,1,1,3,1,1,1
```


If both files should be **combined into one csv file** the `--period=aggregated` flag can be used:

```
$ keptn-usage-stats --folder=./example_payloads --granularity=service --export=csv --period=aggregated

---
Timeframe: 2020-09-23 14:02:42 +0000 UTC - 2020-06-21 02:41:45 +0000 UTC
---

Overall: Keptn
- Executions: 6
-------------------------------------------------
- gatekeeper-service:            2       sh.keptn.event.approval.finished
- gatekeeper-service:            2       sh.keptn.event.approval.triggered
- argo-service:                  1       sh.keptn.event.deployment.finished
- helm-service:                  1       sh.keptn.event.deployment.finished

Project: Keptn > my-project
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- argo-service:                  1       sh.keptn.event.deployment.finished

Service: Keptn > my-project > users
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- argo-service:                  1       sh.keptn.event.deployment.finished

Project: Keptn > sockshop
- Executions: 3
-------------------------------------------------
- helm-service:                  1       sh.keptn.event.deployment.finished
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered

Service: Keptn > sockshop > carts
- Executions: 3
-------------------------------------------------
- gatekeeper-service:            1       sh.keptn.event.approval.finished
- gatekeeper-service:            1       sh.keptn.event.approval.triggered
- helm-service:                  1       sh.keptn.event.deployment.finished


```

**stats.csv**
```
Timeframe,Overall: Keptn,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished),helm-service (sh.keptn.event.deployment.finished),Project: Keptn > my-project,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished),Service: Keptn > my-project > users,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished),Project: Keptn > sockshop,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),helm-service (sh.keptn.event.deployment.finished),Service: Keptn > sockshop > carts,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),helm-service (sh.keptn.event.deployment.finished)
2020-09-23 14:02:42 +0000 UTC - 2020-06-21 02:41:45 +0000 UTC,6,2,2,1,1,3,1,1,1,3,1,1,1,3,1,1,1,3,1,1,1
```


