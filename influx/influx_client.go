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

	"gerrit.o-ran-sc.org/r/qp-aiml/data"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/spf13/viper"
)

const (
	INFLUX_MEASUREMENT_NAME = "ricIndication_cellMetrics"
	INFLUX_FIELD_NAME       = "Cell Metrics"
)

type InfluxConfigs struct {
	Url    string
	Token  string
	Bucket string
	Org    string
}

type InfluxClient struct {
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

func CreateInfluxClient() InfluxClient {
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

	return InfluxClient{
		InfluxConfig,
		client,
	}
}

func (c *InfluxClient) RetrieveCellMetrics(ueid string) ([]data.CellMetricsEntry, error) {

	queryApi := c.client.QueryApi(c.influx.Org)

	query := `from(bucket:"` + c.influx.Bucket + `") 
	|> range(start:-1d)
	|> filter(fn: (r)=>r["_measurement"] == "` + INFLUX_MEASUREMENT_NAME + `")
	|> filter(fn: (r) => r["_field"] == "` + INFLUX_FIELD_NAME + `")`

	result, err := queryApi.Query(context.Background(), query)
	if err != nil {
		xapp.Logger.Error("failed to query (%s) : %s", query, err)
		return nil, err
	}

	var cellMetricsEntries []data.CellMetricsEntry

	for result.Next() {
		var cellMetricsEntry data.CellMetricsEntry
		err := json.Unmarshal([]byte(result.Record().String()), &cellMetricsEntry)
		if err != nil {
			xapp.Logger.Error("failed to unmarshal : %s", err)
			return nil, err
		}
		cellMetricsEntries = append(cellMetricsEntries, cellMetricsEntry)
	}
	return cellMetricsEntries, nil
}
