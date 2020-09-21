# Statistics Service

This service provides usage statistics about a Keptn installation.

## Deploy in your Kubernetes cluster

To deploy the current version of the *shipyard-controller* in your Keptn Kubernetes cluster, use the files `deploy/pvc.yaml` and `deploy/service.yaml` from this repository and apply it.

```console
kubectl apply -f deploy/service.yaml
```

## Delete in your Kubernetes cluster

To delete a deployed *shipyard-controller*, use the files `deploy/pvc.yaml` and `deploy/service.yaml` from this repository and delete the Kubernetes resources:

```console
kubectl delete -f deploy/service.yaml
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

To retrieve usage statistics for a certain time frame, you need to provide the Unix timestamps for the start and end of the time frame.
E.g.:

```
http://localhost:8080/v1/statistics?from=1600656105&to=1600696105
```

cURL Example:

```
curl -X GET "http://localhost:8084/v1/statistics?from=1600656105&to=1600696105" -H "accept: application/json"
```

### Configuring the service

By default, the service aggregates data with a granularity of 30 minutes. Whenever this period has passed, the service will create
a new entry in the Keptn-MongoDB within the Keptn cluster. If you would like to change how often statistics are stored, you can set the 
variable `AGGREGATION_INTERVAL_SECONDS` to your desired value.
