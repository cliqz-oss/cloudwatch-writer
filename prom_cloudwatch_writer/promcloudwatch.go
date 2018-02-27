// prom_cldwatch_writer implements http server listner and a batch upload
// mechanism to cloudwatch. AWS credentials as read in aws-sdk-go, from either
// one of Environment Variable|awsconfig file|IAM role
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

var (
	debug = false
)

func debugPrint(msg string) {
	if debug {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
}

// StartMetricExporter starts listening on serverAddr and will export metrics posted to Cloudwatch
// with namespace and region awsRegion
func StartMetricExporter(serverAddr, namespace, awsRegion string, verbose bool) error {
	debug = verbose
	conn := cloudwatch.New(session.New(), &aws.Config{Region: aws.String(awsRegion)})
	tsQueue := make(chan *prompb.TimeSeries, 10)
	go writeToCloudWatch(conn, tsQueue, namespace)

	debugPrint("listening on: " + serverAddr)
	return runHTTPServer(serverAddr, tsQueue)
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
		datum.SetTimestamp(time.Unix(0, sample.Timestamp*1e6))
		datum.SetValue(sample.Value)
		datumList = append(datumList, datum)
	}

	return datumList, nil
}

func writeToCloudWatch(conn *cloudwatch.CloudWatch, tsQueue <-chan *prompb.TimeSeries, namespace string) {
	ticker := time.NewTicker(5 * time.Second)
	datums := []*cloudwatch.MetricDatum{}

	for {
		select {
		case ts := <-tsQueue:
			datumList, err := getMetricDatum(ts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error converting metrics: %v \n", err)
				break
			}

			if len(datumList) <= 0 {
				fmt.Fprintf(os.Stderr, "error: emtpy datum!")
				break
			}

			for _, d := range datumList {
				datums = append(datums, d)
			}

		case <-ticker.C:
			if len(datums) == 0 {
				break
			}

			debugPrint(fmt.Sprintf("sending %d datapoints to cw", len(datums)))

			metricData := &cloudwatch.PutMetricDataInput{
				Namespace:  &namespace,
				MetricData: datums,
			}

			debugPrint(fmt.Sprintf("writing to cw: %v\n", metricData))
			_, err := conn.PutMetricData(metricData)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: writing to cw: %v\n", err)
			}

			// reset datum to send
			datums = []*cloudwatch.MetricDatum{}
		}
	}
}

func runHTTPServer(addr string, tsQueue chan<- *prompb.TimeSeries) error {
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			debugPrint("request failed: " + err.Error())
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			debugPrint("request failed: " + err.Error())
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			debugPrint("unmarshal failed: " + err.Error())
			return
		}

		for _, ts := range req.Timeseries {
			tsQueue <- ts
		}
	})

	return http.ListenAndServe(addr, nil)
}
