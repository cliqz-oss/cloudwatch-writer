# Prometheus Cloudwatch Remote Writer

cloudwatch-writer is a small web server compatible with [Prometheus Remote Writer Endpoint](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<remote_write>) to export Prometheus metrics to cloudwatch.

## Known problems
1. Cloudwatch only allows at most 10 dimensions for a metrics, metrics with more
   than 10 dimensions is ignored
2. Only simple metrics are tested at the moment.

## Usage:
1. AWS credentials can be passed as environment variables or config files or as IAM roles.
2. invoke the program

```
./cloudwatch-writer --namespace MyNamespace --region us-east-1
```

3. configure Prometheus to send remote write instances to the program. This can be achieved by adding following (sample) to the Prometheus config file

```
remote_write:
  - url: "http://<ip addr>:1234/receive"
```
