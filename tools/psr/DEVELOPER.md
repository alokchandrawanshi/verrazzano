# PSR Developer Guide

This document describes how to develop PSR workers and scenarios that can be used to test specific Verrazzano areas. 
Following is a summary of the steps needed:

1. Get familiar with the PSR tool, run the example and some scenarios.
2. Decide what component you want to test.
3. Decide what you want to test for your first scenario.
4. Decide what workers you need to implement your scenario use cases.
5. Implement a single worker and test it using Helm.
6. Create or update a scenario that includes the worker.
7. Test the scenario using the PSR CLI (psrctl).
8. Repeat steps 5-7 until the scenario is complete.
9. Update the README with your worker information.

## Prerequisites
- Read the [Verrazzano PSR README](./README.md)  to get familiar with the PSR concepts and structure of the source code.
- A Kubernetes cluster with Verrazzano installed (full installation or the components you are testing).

## PSR Areas
Workers are organized into areas, where each area typically maps to one or more Verrazzano backend components, but that isn't always
the case as shown with HTTP workers.  You can see the workers in the [workers](./backend/workers) package.  
PSR scenarios are also grouped into areas.

The following area names are used in the source code and YAML configuration.
They are not exposed in metrics names, rather each `worker.go` file specifies the metrics prefix, which is the long name.  
For example, the OpenSearch worker uses the metric prefix `opensearch`

1. argo - Argo
2. oam - OAM applications, Verrazzano application operator
3. cm - cert-manager
4. cluster - Verrazzano Cluster operator, multicluster
5. coh - Coherence
6. dns - ExternalDNS
7. jaeger - Jaeger
8. kc - Keycloak
9. http - HTTP tests
10. istio - Istio, Kiali
11. mysql - MySQL
12. nginx - NGINX Ingress Controller, AuthProxy
13. ops - OpenSearch, OpenSearchDashboards, Fluentd, VMO
14. prom - Prometheus stack, Grafana
15. rancher - Rancher
16. velero - Velero
17. wls - WebLogic

## Developing a worker
As mentioned in the README, a worker is the code that implements a single use case. For example, a worker might continuously
scale OpenSearch in and out.  The `DoWork` function is the code that actually does the work for one loop iteration, and 
is called repeatedly by the `runner`.  DoWork does whatever it needs to do to perform work, this includes blocking calls or 
condition checks.

### Worker Tips
Here is some important information to know about workers, much of it is repeated in the README.

1. Worker code runs in a backend pod.
2. The same backend pod has the code for all the workers, but only one worker is executing.
3. Workers can have multiple threads doing work (scale up).
4. Workers can have multiple replicas (scale out).
5. Workers are configured using environment variables.
6. Workers should only do one thing (e.g., query OpenSearch).
7. All worker should emit metrics.
8. Workers must wait for their dependencies before doing work (e.g., Verrazzano CR ready).
9. Worker `DoWork` function is called repeatedly in a loop by the `runner`.
10. Some workers must be run in an Istio enabled namespace (depends on what the worker does).
11. A Worker might need additional Kubernetes resources to be created (e.g., AuthorizationPolicies).
12. Workers can be run as Kubernetes deployments or OAM apps (default), this is specified at Helm install.
13. All workers run as cluster-admin.

### Worker Chart and Overrides
Workers are deployed using Helm where there is a single Helm chart for all workers along with area specific Helm subcharts.
Each worker specifies the value overrides in a YAML file, such as the environment variables needed to configure
worker. If an area specific subchart is needed, then it must be enabled in the override file.

The worker override YAML file is in manifests/usecases/<area>/<worker>.yaml.  The only environment variable required is
the `PSR_WORKER_TYPE`. For example, [usecases/opensearch/getlogs.yaml](./manifests/usecases/opensearch/getlogs.yaml)

```
global:
  envVars:
    PSR_WORKER_TYPE: ops-getlogs
    
# activate subchart
opensearch:
  enabled: true

```

### Sample MySQL worker
To make this section easier to follow, we will describe creating a new MySQL worker that queries the MySQL database.  
In general, when creating a worker, it is easiest to just copy an existing worker that does the same type of action (e.g., scale)
and modify it as needed for your component.  When it makes sense, common code should be factored out and reused by multiple workers.

### Creating a worker skeleton
Following are the first steps to implement a worker:

1. Add a worker type named `WorkerTypeMysqlQuery = mysql-query` to [config.go](./backend/config/config.go).
2. Create a package named `mysql` in package [workers](./backend/workers).
3. Create a file `query.go` in the `mysql` package and do the following:
   1. Stub out the [worker interface](./backend/spi/worker.go) implementation in `query.go`  You can copy the ops getlogs worker as a starting point.
   2. Change the const metrics prefix to `metricsPrefix = "mysql_query"`.
   3. Rename the `NewGetLogsWorker` function to `NewQueryWorker`.
   4. Change the `GetWorkerDesc` function to return information about the worker.
   6. Change the DoWork function to  `fmt.Println("hello mysql query worker")`.
4. Add your worker case to the `getWorker` function in [manager.go](./backend/workmanager/manager.go).
5. Add a directory named `mysql` to [usecases](./manifests/usecases).
6. Copy [usecases/opensearch/getlogs.yaml](./manifests/usecases/opensearch/getlogs.yaml) to a file named `usecases/mysql/query.yaml`.
7. Edit query.yaml:
   1. change `PSR_WORKER_TYPE: ops-getlogs` to `PSR_WORKER_TYPE: mysql-query`.
   2. remove the opensearch-authpol section.

### Testing the worker skeleton
This section shows how to test the new worker in a Kind cluster.

1. Test the example worker first by building the image, loading it into the cluster and running the example worker:
   1. `make run-example-k8s`.
   2. take note of the image name:tag that is used with the --set override, for example the output might show this:
      1. helm upgrade --install psr manifests/charts/worker --set appType=k8s --set imageName=ghcr.io/verrazzano/psr-backend:local-4210a50.
2. kubectl get pods to see the example worker, look at the pod logs to make sure it is logging.
3. Delete the example worker:
   1. `helm delete psr`.
4. Run the mysql worker with the newly built image, an example image tag is shown below:
   1. `helm install psr manifests/charts/worker -f manifests/usescases/mysql/query.yaml --set appType=k8s --set imageName=ghcr.io/verrazzano/psr-backend:local-4210a50`
5. Look at the PSR mysql worker pod and make sure that it is logging `hello mysql query worker`.
6. Delete the mysql worker:
   1. `helm delete psr`.

### Add worker specific charts
To function properly, certain workers need additional Kubernetes resources to be created.  Rather than having the worker create the
resources at runtime, you can use a subchart to create them. The subchart will be shared by all workers in an area.  
Since the MySQL query worker needs to access MySQL directly within the cluster, it will need an Istio AuthorizationPolicy,
just like the OpenSearch workers do.  This section will show how to add the chart and use it in the use case YAML file.

1. Create a new subchart called `mysql`:
   1. copy the opensearch chart from [manifests/charts/worker/charts/opensearch](./manifests/charts/worker/charts/opensearch) to 
[manifests/charts/worker/charts/mysql](./manifests/charts/worker/charts/mysql).
   2. create the authorizationpolicy.yaml file with the correct policy to access MySQL.
   3. Delete the existing opensearch policy yaml files.
2. Edit the [worker Chart.yaml](./manifests/charts/worker/Chart.yaml) file and add a dependency for the mysql chart.
```
dependencies:
  - name: mysql
    repository: file://../mysql
    version: 0.1.0
    condition: mysql.enabled
```
3.Edit the [worker Chart.yaml](./manifests/charts/worker/Chart.yaml) file and add the following section:
```
# activate subchart
mysql:
  enabled: false
```
4. Edit [usecases/mysql/query.yaml](./manifests/usecases/mysql/query.yaml) and add the following section:
```
# activate subchart
mysql:
  enabled: true
```
5. You will need to install the chart in an Istio enabled namespace.
6. Test the chart in an Verrazzano installation using the same Helm command as previously, but also specify the namespace:
   1. `helm install psr manifests/charts/worker -n myns -f manifests/usescases/mysql/query.yaml --set appType=k8s --set imageName=ghcr.io/verrazzano/psr-backend:local-4210a50`.

### Add metrics to worker
Worker metrics are very important because they let us track the progress and health of a worker.  Before implementing 
the `DoWork` and `PreconditionsMet` functions, you should get metrics working.  The reason is that you will be able to 
easily test your metrics by running your worker in an IDE, then opening up your browser to http://localhost:9090/metrics.  
Once you implement the real worker code (`DoWork`), you might need to run in an Istio enabled namespace and will need 
to use Prometheus or Grafana to see the metrics.

The [runner](./backend/workmanager/runner.go) also emits metrics such as loop count, so you don't need to emit the same metrics.

1. Modify the `workerMetrics` struct to add the metrics that the worker will emit.
2. Modify the `NewQueryWorker` function to specify the metrics descriptors:
   1. Use a CounterValue metric if the value can never go down, otherwise use GaugeValue or some other metric type.
   2. Don't specify the worker type prefix in the name field, that is automatically added to the metric name.
3. Modify the `GetMetricList` function returning the list of metrics.
4. Modify DoWork to update the metrics as work is done:
   1. You might have some metrics that you cannot implement until the full DoWork code is done.
   2. Metric access must be thread-safe, use the atomic package like the other worker.
5. Test the worker using the Helm chart.
6. Access the Prometheus console and query the metrics.

### Implement the remainder of the worker code
Implement the remaining worker code in `query.go`, specifically `PreconditionsMet` and `DoWork` Note that the query worker
doesn't need a Kubernetes client since it knows the MySQL service name. If your worker needs
to call the Kubernetes API server, then use the [k8sclient](./backend/pkg/k8sclient) package.  See how the
OpenSearch [getlogs](./backend/workers/opensearch/scale/scale.go) worker uses ` k8sclient.NewPsrClient`.

1. Implement NewQueryWorker to create the worker instance.
2. Change function GetEnvDescList to return configuration environment variables that the worker needs:
   1. See the OpenSearch [getlogs](./backend/workers/opensearch/scale/scale.go) worker for an example.
3. Implement DoWork. This method should not log, but if it really needs to log, then use the throttled Verrazzano logging,
   such as Progress or ErrorfThrottled.
4. Test the worker using the Helm chart.

**NOTE** The same worker instance is shared across all worker threads.  There is currently no state per worker.  Workers
that keep state, such as the scaling worker, normally only run in a single thread.

## Creating a scenario
A scenario is a collection of use cases with a curated configuration, that are run concurrently.  Typically,
you should restrict the scenario use cases to a single area, but that is not a strict requirement.  You can run multiple
scenarios concurrently so creating a mixed-area scenario might not be necessary.  If you do decide to create a mixed area scenario,
then create it in a directory called scenario/mixed.

### Scenario files
Scenarios are specified by a scenario.yaml file along with use case override files, one for each use case.
By convention, the files must be in the [<area>/<scenario-name>/scenarios](./manifests/scenarios) directory structured as follows:
```
<area>/<scenario-name>/scenario.yaml
<area>/<scenario-name>usecase-overrides/*
```
Use the long name for the area, e.g. opensearch instead of ops, or cert-manager instead of cm.
For example, the scenario to restart all OpenSearch tiers is in [restart-all-tiers](./manifests/scenarios/opensearch/restart-all-tiers)

### Scenario YAML file
The scenario YAML file describes the scenario, the use cases that comprise the scenario, and the use case overrides for the scenario.
Following is the OpenSearch [restart-all-tiers/scenario.yaml](./manifests/scenarios/opensearch/restart-all-tiers/scenario.yaml).
```
name: opensearch-restart-all-tiers
ID: ops-rat
description: |
  This is a scenario that restarts pods on all 3 OpenSearch tiers simultaneously
usecases:
  - usecasePath: opensearch/restart.yaml
    overrideFile: restart-master.yaml
    description: restarts master nodes
  - usecasePath: opensearch/restart.yaml
    overrideFile: restart-data.yaml
    description: restarts data nodes
  - usecasePath: opensearch/restart.yaml
    overrideFile: restart-ingest.yaml
    description: restarts ingest nodes
```
The front section has the scenario name, ID, and description.  The usecases section has all the use cases that will be run when
the scenario is started.  The `usecasePath` points to the built-in use case override file.  The `overrideFile` specifies the file
in the `usecase-overrides` directory which contains the scenario overrides for that specific use case.

Following is the [scenario override file](./manifests/scenarios/opensearch/restart-all-tiers/usecase-overrides/restart-data.yaml) for restarting the data tier.
```
global:
  envVars:
    PSR_WORKER_TYPE: ops-restart
    OPENSEARCH_TIER: data
    PSR_LOOP_SLEEP: 5s
```
This file specifies that the data tier should be restarted every 5 seconds.  The `PSR_WORKER_TYPE` is not really needed here, since it
is already in the use case file at [usecases/opensearch/restart.yaml](./manifests/usecases/opensearch/restart.yaml), however, it
doesn't hurt to have it for documentation purposes.

## Running a scenario
Scenarios are run by the PSR command line interface, `psrctl`.  The source code [manifests](./manifests) directory contains all
of the helm charts, use cases overrides, and scenario files.  These manifests files are built into the psrctl binary and accessed
internally at runtime, so the psrctl binary is self-contained and there is no need for the user to provide external files.  However,
you can override the scenario directory at runtime with the `-d` flag.  This allows you to modify and test scenarios without having
to rebuild `psrctl`.  See `psrctl` help for details.

Scenarios always use OAM to deploy the workers, Kubernetes Deployments are not an option at this time.

### Specifying the backend image
If you build `psrctl` using make, the image tag is derived from the last commit id.  If that image has not been uploaded to
ghcr.io, you will need to run `make docker-push`.  Since that image is private will need to provide a secret with the ghcr.io
credentials with `psrctl -p`.  If you want to override the image name, use `psrctl -w`.

If you want to use a local image and load it into a kind cluster, they run `make kind-load-image` and specify that image
using `psrctl -w`.  This is the easiest way to develop and test a worker on Kind.

### Updating a running scenario
During development, you may want to update scenario override values while testing the scenario.  Currently, you
can only update with `psrctl -u` but only the `usecase-overrides` files can be changed. If you need to change a chart
or scenario.yaml, then just restart the scenario.

### Sample psrctl commands
This section shows examples to run some built-in scenarios.

**Show the built-in scenarios**
``` 
psrctl explain
```

**Start the OpenSearch scenario ops-s2 in Istio enabled namespace using a custom image**
``` 
kubectl create ns psrtest
kubectl label ns psrtest verrazzano-managed=true istio-injection=enabled
psrctl start -s ops-s2 -n psrtest -w ghcr.io/verrazzano/psr-backend:local-c2e911e
```

**Show the running scenarios in namespace psrtest, then across all namespaces**
``` 
psrctl list -n psrtest 
psrctl list -A
```

**Update the running OpenSearch scenario ops-s2 with external scenario override files**
``` 
psrctl update -s ops-s2 -n psrtest -d ~/tmp/my-ops-s2
```

**Stop the running OpenSearch scenario ops-s2**
``` 
psrctl stop -s ops-s2 -n psrtest
```

## Source Code
The source code is organized into backend code, psrctl code, and manifest files.

### Backend code
The  [backend](./backend) directory has the backend code which consists of the following packages:
* config - configuration code
* metrics - metrics server for metrics generated by the workers
* osenv - package that allows workers to specify default and required env vars
* pkg - various packages needed by the workers
* spi - the worker interface
* workers - the various workers
* workmanager - the worker manager and runner

### Manifests
The [manifests/charts](./manifests/charts)  directory has the Helm charts. There is a single worker chart for using
either OAM or plain Kubernetes resources to deploy the backend.  The default is OAM, use
deploy without OAM, use `--set appType=k8s`.  There is also a subchart for each area that requires one.

The [manifests/usecases](./manifests/usecases) directory has the Helm override files for every use
case. These files must contain the configuration, as key:value pairs, required by the worker.
The usecases are organized by area.

The [manifests/scenarios](./manifests/scenarios) directories for each scenario. Each scenario director
contains the scenario.yaml file for the scenario, along with use case override files.
The scenarios are organized by area.

### Psrctl
The [psrct](./psrctl) directory contains the command line interface along with support packages.  One
thing to note is that the [embed.go](./embed.go) file is needed by the psrctl code to access the built-in
manifests.  This file needs to be in the parent directory of psrctl.


## Summary
This document has all the information needed to create new workers and scenarios for any Verrazzano component.
When creating workers, it is easiest to use Helm to deploy and test the worker. Always start with a single worker
thread, then test with multiple threads and replicas. 

When using Helm directly, you can deploy your worker as a Kubernetes Deployment or as an OAM application.  
If you use a Deployment, you can start testing your stubbed-out worker without Verrazzano installed.  
When you get further and need Verrazzano dependencies or want the worker metrics scraped, you can switch to OAM.  
Your worker needs to run in the target cluster, or the worker metrics won't get scraped.

If your worker needs to access resource in the mesh, like OpenSearch, you will need to create AuthorizationPolicies 
(via a subchart) and will need to deploy your worker to an Istio enabled namespace.  The existing OpenSearch 
workers have this requirement so you can use one of them as a starting point. Make sure you use Prometheus to 
test your worker metrics.  Once the worker is running, you can create any custom scenario using YAML files as described earlier.  
Finally, use the `psrctl` CLI to run and test your scenarios.
