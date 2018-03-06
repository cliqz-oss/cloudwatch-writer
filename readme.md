# Prometheus Cloudwatch Remote Writer

[![Build Status](https://travis-ci.org/cliqz-oss/cloudwatch-writer.svg?branch=master)](https://travis-ci.org/cliqz-oss/cloudwatch-writer)

cloudwatch-writer is a small web server compatible with [Prometheus Remote Writer Endpoint](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#<remote_write>) to export Prometheus metrics to cloudwatch.

Prometheus is much powerful than cloudwatch when it comes to working with metrics, but cloudwatch integrates well with other aws services. This program bridges the gap.

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


### All Options
```
Usage:
  cloudwatch-writer [flags]

Flags:
      --debug               enable debug output
  -h, --help                help for cloudwatch-writer
  -n, --namespace string    namespace for cloudwatch metrics
      --region string       aws region
  -s, --serveraddr string   server address listen for prometehus remote writes (default "0.0.0.0:1234")
```

TODO:
 - [ ] tests for more complex metrics.
