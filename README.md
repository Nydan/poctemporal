# POC Temporal

This repository is a sample worker and http server for triggering Temporal
workflow. The application implement TLS for connecting to Temporal cluster.

## Separated Worker and Workflow trigger

To run the POC:
1. Generate certificate by following [tls-simple](https://github.com/temporalio/samples-server/tree/main/tls/tls-simple).
1. Copy the generated `/certs` folder to this root repo.
1. Run the `docker-compose` from [tls-simple](https://github.com/temporalio/samples-server/tree/main/tls/tls-simple).
1. Create a namespace by running this command:

    ```
    docker exec tls-simple_temporal-admin-tools_1 tctl --namespace {namespace} namespace register
    ```
    It run the `tctl` command from the cluster with TLS. Replace the 
    `{namespace}` value and change the configuration in `development.yaml`.
1. Run the worker:
    ```
    go run worker/main.go
    ```
1. Run the http server:
    ```
    go run start/main.go
    ```
1. Try to create API request:
    ```
    curl --location --request POST 'localhost:8084' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "Amount": 1000,
        "UserID": "nydan"
    }'
    ```

## Apps with http server and workers

Running binary in `./apps` already run the workers and the http server.

```
go run apps/main.go
```
