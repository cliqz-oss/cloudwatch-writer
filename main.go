// prom-cloudwatch-remote-writer writes incoming metrics from prometheus
// (configured using remote_write config) to cloudwatch.
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	//	ServerAddr = "0.0.0.0:1234"
	AwsRegion = "us-east-1"
	Namespace = "Fetcher"
)

func main() {
	Conn := cloudwatch.New(session.New(), &aws.Config{Region: aws.String(AwsRegion)})
	writeToCloudWatch(Conn)
}

func writeToCloudWatch(conn *cloudwatch.CloudWatch) {
	// create and add a dummy metrics

	dims := map[string]string{
		"dim1": "cat",
		"dim2": "dog",
	}

	metricDatum := &cloudwatch.MetricDatum{}
	metricDatum.SetMetricName("customuploadmetrics")
	dimens := []*cloudwatch.Dimension{}
	for k, v := range dims {
		d := &cloudwatch.Dimension{}
		d.SetName(k)
		d.SetValue(v)
		dimens = append(dimens, d)
	}

	metricDatum.SetDimensions(dimens)
	metricDatum.SetTimestamp(time.Unix(1519644379, 0))
	metricDatum.SetValue(0.050107)

	fmt.Println(metricDatum.String())
	metricData := cloudwatch.PutMetricDataInput{
		Namespace:  &Namespace,
		MetricData: []*cloudwatch.MetricDatum{metricDatum},
	}

	out, err := conn.PutMetricData(&metricData)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}

func runHTTPServer(addr string, metricsQueue chan<- model.Metric) {

	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, ts := range req.Timeseries {
			m := make(model.Metric, len(ts.Labels))
			for _, l := range ts.Labels {
				m[model.LabelName(l.Name)] = model.LabelValue(l.Value)
			}
			fmt.Println(m)
			for _, s := range ts.Samples {
				fmt.Printf("  %f %d\n", s.Value, s.Timestamp)
			}
		}
	})

	http.ListenAndServe(addr, nil)
}
