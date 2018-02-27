// prom-cloudwatch-remote-writer writes incoming metrics from prometheus
// (configured using remote_write config) to cloudwatch.
package prom_cldwatch_writer

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
	"os"
	"time"
)

// StartMetricExporter starts listening on serverAddr and will export metrics posted to Cloudwatch
// with namespace and region awsRegion
func StartMetricExporter(serverAddr, namespace, awsRegion string) {
	conn := cloudwatch.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})
	tsQueue := make(chan *prompb.TimeSeries)
	go writeToCloudWatch(conn, tsQueue)
	fmt.Fprintf(os.Stderr, "listening on: %s\n", serverAddr)
	fmt.Fprintf(os.Stderr, "%v", runHTTPServer(serverAddr, tsQueue))
}

func getMetricDatum(ts *prompb.TimeSeries) ([]*cloudwatch.MetricDatum, error) {
	if (len(ts.Labels)) > 10 {
		return nil, fmt.Errorf("cloudwatch only allow 10 dimensions. got: %v", ts.Labels)
	}

	m := make(model.Metric, len(ts.Labels))
	for _, l := range ts.Labels {
		m[model.LabelName(l.Name)] = model.LabelValue(l.Value)
	}

	mName, ok := m[model.MetricNameLabel]
	if !ok {
		mName = "unnamed"
	}

	// get extra dimensions
	dims := []*cloudwatch.Dimension{}
	for label, value := range m {
		if label != model.MetricNameLabel {
			d := &cloudwatch.Dimension{}
			d.SetName(fmt.Sprint(label))
			d.SetValue(fmt.Sprint(value))
			dims = append(dims, d)
		}
	}

	datumList := []*cloudwatch.MetricDatum{}

	for _, sample := range ts.Samples {
		datum := &cloudwatch.MetricDatum{}
		datum.SetMetricName(fmt.Sprint(mName))
		datum.SetDimensions(dims)

		datum.SetTimestamp(time.Unix(0, sample.Timestamp))
		datum.SetValue(sample.Value)
		datumList = append(datumList, datum)
	}

	return datumList, nil
}

func writeToCloudWatch(conn *cloudwatch.CloudWatch, tsQueue <-chan *prompb.TimeSeries) {
	for ts := range tsQueue {
		datumList, err := getMetricDatum(ts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error converting metrics: %v", err)
			continue
		}
		fmt.Println(datumList)
	}
}

func runHTTPServer(addr string, tsQueue chan<- *prompb.TimeSeries) error {
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
			tsQueue <- ts
		}
	})

	return http.ListenAndServe(addr, nil)
}
