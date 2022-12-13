/*
==================================================================================
      Copyright (c) 2022 Samsung Electronics Co., Ltd. All Rights Reserved.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

         http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

==================================================================================
*/

package influx

import (
	"context"
	"encoding/json"
	"os"

	"gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/data"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/viper"
)

const (
	INFLUX_MEASUREMENT_NAME = "ricIndication_cellMetrics"
	INFLUX_FIELD_NAME       = "Cell Metrics"

	ENV_INFLUX_QUERY_START = "INFLUX_QUERY_START"
	ENV_INFLUX_QUERY_STOP  = "INFLUX_QUERY_STOP"
)

type InfluxConfigs struct {
	Url    string
	Token  string
	Bucket string
	Org    string
}

//go:generate mockgen -package mocks -source=$GOFILE -destination=./mocks/mock_$GOFILE
type InfluxDBCommand interface {
	RetrieveCellMetrics() ([]data.CellMetricsEntry, error)
}

type InfluxDB struct {
	influx InfluxConfigs
	client influxdb2.Client
}

func init() {
	viper.SetEnvPrefix("INFLUX")
	viper.BindEnv("URL")
	viper.BindEnv("TOKEN")
	viper.BindEnv("BUCKET")
	viper.BindEnv("ORG")
}

func CreateInfluxDB() InfluxDB {
	var InfluxConfig InfluxConfigs
	err := viper.Unmarshal(&InfluxConfig)
	if err != nil {
		xapp.Logger.Error("failed to Unmarshal InfluxConfigs")
	}
	out, err := json.Marshal(InfluxConfig)
	if err != nil {
		xapp.Logger.Error("failed to json.Marshal InfluxConfigs")
	}
	xapp.Logger.Debug("InfluxConfig : %s", out)

	client := influxdb2.NewClientWithOptions(InfluxConfig.Url, InfluxConfig.Token, influxdb2.DefaultOptions().SetBatchSize(20))

	return InfluxDB{
		InfluxConfig,
		client,
	}
}

func (c *InfluxDB) getQueryRange() string {
	start := os.Getenv(ENV_INFLUX_QUERY_START)
	stop := os.Getenv(ENV_INFLUX_QUERY_STOP)

	if start == "" {
		xapp.Logger.Error("invalid query range !")
		return ""
	}

	if stop == "" {
		return `|> range(start: ` + start + `)`
	}
	return `|> range(start: ` + start + `, stop: ` + stop + `)`
}

func (c *InfluxDB) RetrieveCellMetrics() ([]data.CellMetricsEntry, error) {

	queryAPI := c.client.QueryAPI(c.influx.Org)

	queryRange := c.getQueryRange()

	if queryRange == "" {
		return nil, nil
	}

	query := `from(bucket:"` + c.influx.Bucket + `") 
	` + queryRange + `
	|> filter(fn: (r)=>r["_measurement"] == "` + INFLUX_MEASUREMENT_NAME + `")
	|> filter(fn: (r) => r["_field"] == "` + INFLUX_FIELD_NAME + `")`

	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		xapp.Logger.Error("failed to query (%s) : %s", query, err)
		return nil, err
	}

	var cellMetricsEntries []data.CellMetricsEntry

	for result.Next() {
		var cellMetricsEntry data.CellMetricsEntry
		err := json.Unmarshal([]byte(result.Record().Value().(string)), &cellMetricsEntry)
		if err != nil {
			xapp.Logger.Error("failed to unmarshal : %s", err)
			return nil, err
		}
		cellMetricsEntries = append(cellMetricsEntries, cellMetricsEntry)
	}

	return cellMetricsEntries, nil
}
