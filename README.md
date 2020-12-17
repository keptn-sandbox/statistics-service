# Statistics Service

This service provides usage statistics about a Keptn installation.

## Compatibilty Matrix

| Keptn Version    | [Statistics Service](https://hub.docker.com/r/keptnsandbox/statistics-service/tags?page=1&ordering=last_updated) | Kubernetes Versions                      |
|:----------------:|:----------------------------------------:|:----------------------------------------:|
|       0.7.1      | keptnsandbox/statistics-service:0.1.0    | 1.14 - 1.19                              |
|       0.7.2      | keptnsandbox/statistics-service:0.1.1    | 1.14 - 1.19                              |
|       0.7.3      | keptnsandbox/statistics-service:0.2.0    | 1.14 - 1.19                              |


## Deploy in your Kubernetes cluster

Please note that the installation of the **statistics-service** differs slightly, depending on your installed Keptn version. Depending on your installed Keptn version, please follow the instructions below. 

### For Keptn versions < 0.8.0

To deploy the current version of the *statistics-service* in your Keptn Kubernetes cluster, use the file `deploy/service.yaml` from this repository and apply it.

```console
kubectl apply -f deploy/service.yaml -n keptn
```

### For Keptn versions >= 0.8.0

To deploy the current version of the *statistics-service* in your Keptn Kubernetes cluster, use the file `deploy/service_keptn_080.yaml` from this repository and apply it.

```console
kubectl apply -f deploy/service_keptn_080.yaml -n keptn
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

* To build the CLI locally: 
```
go build -o keptn-usage-stats
```

* How to use the tool: 

```
keptn-usage-stats --help
```


```
Generates an overview of Keptn usage statistics, based on a set of input files provided to the command. Example:

keptn-usage-stats
   --folder=./usage-statistics-xyz
   --period=separated
   --granularity=overall,project
   --includeEvents=deployment-finished,tests-finished,evaluation-done
   --includeServices=all

Usage:
  keptn-usage-stats [flags]

Flags:
      --excludeProjects string   List of project names that are excluded from the Summary
      --export string            The format to export the statistics, supported are [json, csv] (default "json")
  -f, --folder string            The folder containing the JSON files exported from the statistics-service
  -g, --granularity string       The level of details, list of [overall, project, service], default is 'overall' (default "overall")
  -h, --help                     help for keptn-usage-stats
      --includeEvents string     List of events that define an automation unit, default is 'all' (default "all")
      --includeServices string   List of Services that define an automation unit, default is 'all' (default "all")
      --includeTriggers string   List of sequence triggers: [configuration-change, problem.open, evaluation-started] - supported with Keptn >0.8 (default "all")
  -o, --output string            The name of the output file (default "stats")
  -p, --period string            The period under consideration, one option of: [separated, aggregated] (default "separated")
      --separator string         The separator used for the CSV exporter, allowed values are ',' or ';' (default ",")
```

**Note:** The `--includeTriggers` flag is not supported yet, but will be implemented with Keptn 0.8. 

### Examples

#### Example A

The following command will create a single CSV file with a row for each statistics files located in the directory `./example_payloads`:

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

**stats.csv**
```
Timeframe,Overall: Keptn,gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),helm-service (sh.keptn.event.deployment.finished),argo-service (sh.keptn.event.deployment.finished),Project: Keptn > sockshop,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished),Service: Keptn > sockshop > carts,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished),Project: Keptn > my-project,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished),Service: Keptn > my-project > users,argo-service (sh.keptn.event.deployment.finished),gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered)
2020-09-21 02:41:45 +0000 UTC - 2020-09-23 14:02:42 +0000 UTC,3,1,1,1,,3,1,1,1,3,1,1,1,,,,,,,,
2020-06-21 02:41:45 +0000 UTC - 2020-06-23 14:02:42 +0000 UTC,3,1,1,,1,,,,,,,,,3,1,1,1,3,1,1,1

```

#### Example B

If both files should be **combined into one row (of a CSV)** the `--period=aggregated` flag can be used:

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
Timeframe,Overall: Keptn,helm-service (sh.keptn.event.deployment.finished),argo-service (sh.keptn.event.deployment.finished),gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),Project: Keptn > sockshop,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),helm-service (sh.keptn.event.deployment.finished),Service: Keptn > sockshop > carts,helm-service (sh.keptn.event.deployment.finished),gatekeeper-service (sh.keptn.event.approval.triggered),gatekeeper-service (sh.keptn.event.approval.finished),Project: Keptn > my-project,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished),Service: Keptn > my-project > users,gatekeeper-service (sh.keptn.event.approval.finished),gatekeeper-service (sh.keptn.event.approval.triggered),argo-service (sh.keptn.event.deployment.finished)
2020-06-21 02:41:45 +0000 UTC - 2020-09-23 14:02:42 +0000 UTC,6,1,1,2,2,3,1,1,1,3,1,1,1,3,1,1,1,3,1,1,1
```


