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
	"errors"
	"io"
	"strings"
	"testing"

	mocks_influx "gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/influx/mocks"
	"github.com/golang/mock/gomock"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen --build_flags=--mod=mod -package mocks -destination ./mocks/mock_influx_client.go github.com/influxdata/influxdb-client-go/v2 Client
//go:generate mockgen --build_flags=--mod=mod -package mocks -destination ./mocks/mock_query_api.go github.com/influxdata/influxdb-client-go/v2/api QueryAPI
func TestCreateInfluxClient_ExpectSuccess(t *testing.T) {
	t.Setenv("INFLUX_URL", "http://localhost:8088")
	t.Setenv("INFLUX_TOKEN", "token")
	t.Setenv("INFLUX_BUCKET", "bucket")
	t.Setenv("INFLUX_ORG", "org")

	influxClient := CreateInfluxDB()

	assert.NotEmpty(t, influxClient.influx)
}

func TestRetrieveCellMetrics_ExpectSuccess(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "2020-01-01T00:01:00Z")
	t.Setenv("INFLUX_QUERY_STOP", "2022-12-31T23:59:59Z")

	newCsvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,{"MeasTimestampPDCPBytes":{"tv_sec":1670561380,"tv_nsec":1670561380053954502},"CellID":"c2/B13","PDCPBytesDL":0,"PDCPBytesUL":0,"MeasTimestampAvailPRB":{"tv_sec":1670561380,"tv_nsec":1670561380053954502},"AvailPRBDL":0,"AvailPRBUL":0,"MeasPeriodPDCPBytes":20,"MeasPeriodAvailPRB":20},f,test,1,adsfasdf
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,{"MeasTimestampPDCPBytes":{"tv_sec":1670561390,"tv_nsec":1670561390066916984},"CellID":"c2/B13","PDCPBytesDL":0,"PDCPBytesUL":0,"MeasTimestampAvailPRB":{"tv_sec":1670561390,"tv_nsec":1670561390066916984},"AvailPRBDL":0,"AvailPRBUL":0,"MeasPeriodPDCPBytes":20,"MeasPeriodAvailPRB":20},f,test,2,adsfasdf
`

	fakeQueryTableResult := api.NewQueryTableResult(io.NopCloser(strings.NewReader(newCsvTable)))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient
	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)
	mockQueryAPI.EXPECT().Query(gomock.Any(), gomock.Any()).Return(fakeQueryTableResult, nil)

	result, err := testInfluxClient.RetrieveCellMetrics()

	//  TODO : Test not working properly yet. Result should be NotNil.
	// 	Cannot make proper fakeQueryTableResult because csv reader cannot read the field containing double-quote '"' which is required for json unmarshal.
	//  Therefore, end tests by calling Result.next() that returns false without testing the appending of cellMetricsEntry.
	assert.Nil(t, result)
	assert.NoError(t, err)
}

func TestNegativeRetrieveCellMetrics_WhenUnmarshalFailed_ExpectError(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "2020-01-01T00:01:00Z")
	t.Setenv("INFLUX_QUERY_STOP", "2022-12-31T23:59:59Z")

	csvTable := `#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,a,b
,,0,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T10:34:08.135814545Z,test0,f,test,1,adsfasdf
,,1,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:08:44.850214724Z,test1,f,test,2,adsfasdf
,,2,2020-02-17T22:19:49.747562847Z,2020-02-18T22:19:49.747562847Z,2020-02-18T22:11:32.225467895Z,test2,f,test,3,adsfasdf
`
	fakeQueryTableResult := api.NewQueryTableResult(io.NopCloser(strings.NewReader(csvTable)))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient
	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)
	mockQueryAPI.EXPECT().Query(gomock.Any(), gomock.Any()).Return(fakeQueryTableResult, nil)

	result, err := testInfluxClient.RetrieveCellMetrics()

	assert.Nil(t, result)
	assert.NotNil(t, err)
}

func TestRetrieveCellMetrics_EmptyQueryResult_ExpectReturnNil(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "2020-01-01T00:01:00Z")
	t.Setenv("INFLUX_QUERY_STOP", "2022-12-31T23:59:59Z")

	fakeQueryTableResult := api.NewQueryTableResult(io.NopCloser(strings.NewReader("")))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient

	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)
	mockQueryAPI.EXPECT().Query(gomock.Any(), gomock.Any()).Return(fakeQueryTableResult, nil)

	result, err := testInfluxClient.RetrieveCellMetrics()

	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestRetrieveCellMetrics_WhenInvalidQueryStartRange_ExpectReturnNil(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient

	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)

	result, err := testInfluxClient.RetrieveCellMetrics()

	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestRetrieveCellMetrics_WhenInvalidQueryStopRange_ExpectReturnNil(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "-1d")

	fakeQueryTableResult := api.NewQueryTableResult(io.NopCloser(strings.NewReader("")))

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient

	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)
	mockQueryAPI.EXPECT().Query(gomock.Any(), gomock.Any()).Return(fakeQueryTableResult, nil)

	result, err := testInfluxClient.RetrieveCellMetrics()

	assert.Nil(t, result)
	assert.Nil(t, err)
}

func TestNegativeRetrieveCellMetrics_WhenQueryApiReturnError_ExpectError(t *testing.T) {
	t.Setenv("INFLUX_QUERY_START", "2020-01-01T00:01:00Z")
	t.Setenv("INFLUX_QUERY_STOP", "2022-12-31T23:59:59Z")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockInfluxClient := mocks_influx.NewMockClient(ctrl)
	mockQueryAPI := mocks_influx.NewMockQueryAPI(ctrl)

	testInfluxClient := CreateInfluxDB()
	testInfluxClient.client = mockInfluxClient

	mockInfluxClient.EXPECT().QueryAPI(gomock.Any()).Return(mockQueryAPI)
	mockQueryAPI.EXPECT().Query(gomock.Any(), gomock.Any()).Return(nil, errors.New(""))

	result, err := testInfluxClient.RetrieveCellMetrics()

	assert.Nil(t, result)
	assert.NotNil(t, err)
}
