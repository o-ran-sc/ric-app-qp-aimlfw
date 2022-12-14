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
package control

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mocks_control "gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/control/mocks"
	"gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/data"
	mocks_influx "gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/influx/mocks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func createPostTestServer(expectedHttpStatus int, expectedBody []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			if req.URL.Path == "/v1/models/qoe-model:predict" {
				rw.Header().Set("Content-Type", "application/x-www-form-urlencoded")
				rw.WriteHeader(expectedHttpStatus)
				rw.Write(expectedBody)
				return
			}
		}
	}))
}

func TestNewControl_ExpectSuccess(t *testing.T) {
	t.Setenv("MLXAPP_HEADERHOST", "qoe-model.kserve-test.example.com")
	t.Setenv("MLXAPP_REQURL", "v1/models/qoe-model:predict")

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	control := NewControl()

	assert.NotEmpty(t, control.mlxAppConfigs)
}

func TestHandleRequestPrediction_ExpectSuccess(t *testing.T) {
	server := createPostTestServer(http.StatusOK, nil)
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Setenv("RIC_MSG_BUF_CHAN_LEN", "256")
	t.Setenv("MLXAPP_HEADERHOST", "qoe-model.kserve-test.example.com")
	t.Setenv("MLXAPP_HOST", strings.Join(strings.Split(server.URL, ":")[:2], ":"))
	t.Setenv("MLXAPP_PORT", strings.Split(server.URL, ":")[2])
	t.Setenv("MLXAPP_REQURL", "v1/models/qoe-model:predict")

	pr, _ := json.Marshal(data.PredictRequest{
		UEPredictionSet: []string{"Car-1"},
	})

	msg := &xapp.RMRParams{
		Payload: pr,
	}

	cellMetricsEntries := []data.CellMetricsEntry{
		{
			MeasTimestampPDCPBytes: data.Timestamp{
				TVsec:  1670561380,
				TVnsec: 1670561380053954502,
			},
			CellID:      "c2/B13",
			PDCPBytesDL: 0,
			PDCPBytesUL: 0,
			MeasTimestampPRB: data.Timestamp{
				TVsec:  1670561380,
				TVnsec: 1670561380053954502,
			},
			AvailPRBDL:     0,
			AvailPRBUL:     0,
			MeasPeriodPDCP: 20,
			MeasPeriodPRB:  20,
		},
	}

	mockInfluxDB := mocks_influx.NewMockInfluxDBCommand(ctrl)
	mockInfluxDB.EXPECT().RetrieveCellMetrics().Return(cellMetricsEntries, nil)

	control := NewControl()
	control.influxDB = mockInfluxDB
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestPrediction(msg)
}

func TestNegativeHandleRequestPrediction_WhenRequestUnmarshalFailed_ExpectReturn(t *testing.T) {
	server := createPostTestServer(http.StatusOK, nil)
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msg := &xapp.RMRParams{}

	control := NewControl()
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestPrediction(msg)
}

func TestNegativeHandleRequestPrediction_WhenInfluxQueryFailed_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pr, _ := json.Marshal(data.PredictRequest{
		UEPredictionSet: []string{"Car-1"},
	})
	msg := &xapp.RMRParams{
		Payload: pr,
	}

	mockInfluxDB := mocks_influx.NewMockInfluxDBCommand(ctrl)
	mockInfluxDB.EXPECT().RetrieveCellMetrics().Return(nil, errors.New(""))

	control := NewControl()
	control.influxDB = mockInfluxDB
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestPrediction(msg)
}

func TestNegativeHandleRequestPrediction_WhenInfluxQueryResultEmpty_ExpectReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pr, _ := json.Marshal(data.PredictRequest{
		UEPredictionSet: []string{"Car-1"},
	})
	msg := &xapp.RMRParams{
		Payload: pr,
	}

	cellMetricsEntries := []data.CellMetricsEntry{}

	mockInfluxDB := mocks_influx.NewMockInfluxDBCommand(ctrl)
	mockInfluxDB.EXPECT().RetrieveCellMetrics().Return(cellMetricsEntries, nil)

	control := NewControl()
	control.influxDB = mockInfluxDB
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestPrediction(msg)
}

func TestNegativeHandleRequestPrediction_WhenResponseStatusBadRequest_ExpectReturn(t *testing.T) {
	server := createPostTestServer(http.StatusBadRequest, nil)
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Setenv("RIC_MSG_BUF_CHAN_LEN", "256")
	t.Setenv("MLXAPP_HEADERHOST", "qoe-model.kserve-test.example.com")
	t.Setenv("MLXAPP_HOST", strings.Join(strings.Split(server.URL, ":")[:2], ":"))
	t.Setenv("MLXAPP_PORT", strings.Split(server.URL, ":")[2])
	t.Setenv("MLXAPP_REQURL", "v1/models/qoe-model:predict")

	pr, _ := json.Marshal(data.PredictRequest{
		UEPredictionSet: []string{"Car-1"},
	})

	msg := &xapp.RMRParams{
		Payload: pr,
	}

	cellMetricsEntries := []data.CellMetricsEntry{
		{
			MeasTimestampPDCPBytes: data.Timestamp{},
			MeasTimestampPRB:       data.Timestamp{},
		},
	}

	mockInfluxDB := mocks_influx.NewMockInfluxDBCommand(ctrl)
	mockInfluxDB.EXPECT().RetrieveCellMetrics().Return(cellMetricsEntries, nil)

	control := NewControl()
	control.influxDB = mockInfluxDB
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestPrediction(msg)
}

func TestConsume_ExpectSuccess(t *testing.T) {
	msg := &xapp.RMRParams{
		Meid: &xapp.RMRMeid{},
	}

	control := NewControl()
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.Consume(msg)
	ret := <-control.rcChan

	assert.NotNil(t, ret)
}

func TestHandleRequestQoEPrediction_ExpectSuccess(t *testing.T) {
	pr, _ := json.Marshal(data.QoePredictionRequest{
		PredictionUE: "12345",
		UEMeasurement: data.UEMeasurementType{
			ServingCellID:            "310-680-200-555002",
			MeasTimestampUEPDCPBytes: "2020-03-18 02:23:18.220",
			MeasPeriodUEPDCPBytes:    20,
			UEPDCPBytesDL:            2500000,
			UEPDCPBytesUL:            1000000,
			MeasTimestampUEPRBUsage:  "2020-03-18 02:23:18.220",
			MeasPeriodUEPRBUsage:     20,
			UEPRBUsageDL:             10,
			UEPRBUsageUL:             30,
		},
		CellMeasurements: []data.CellMeasurement{
			data.CellMeasurement{
				CellID:                 "310-680-220-555001",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            250000,
				PDCPBytesUL:            100000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             30,
				AvailPRBUL:             50,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -90, RSRQ: -13, RSSINR: -2.5},
			},
			data.CellMeasurement{
				CellID:                 "310-680-220-555003",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            200000,
				PDCPBytesUL:            120000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             60,
				AvailPRBUL:             80,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -140, RSRQ: -17, RSSINR: -6},
			},
			data.CellMeasurement{
				CellID:                 "310-680-220-555002",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            190000,
				PDCPBytesUL:            100000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             30,
				AvailPRBUL:             45,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -115, RSRQ: -16, RSSINR: -5},
			},
		},
	})

	qpr, _ := json.Marshal(data.QoePredictionResult{
		Predictions: [][]float32{{0.20054793, 0.2541615}, {0.19025849, 0.24147518}, {0.17816328, 0.2247211}},
	})

	server := createPostTestServer(http.StatusOK, qpr)
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Setenv("RIC_MSG_BUF_CHAN_LEN", "256")
	t.Setenv("MLXAPP_HEADERHOST", "qoe-model.kserve-test.example.com")
	t.Setenv("MLXAPP_HOST", strings.Join(strings.Split(server.URL, ":")[:2], ":"))
	t.Setenv("MLXAPP_PORT", strings.Split(server.URL, ":")[2])
	t.Setenv("MLXAPP_REQURL", "v1/models/qoe-model:predict")

	msg := &xapp.RMRParams{
		Payload: pr,
	}

	control := NewControl()
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestQoEPrediction(msg)
}

func TestNegativeHandleQoERequestPrediction_WhenRequestUnmarshalFailed_ExpectReturn(t *testing.T) {
	server := createPostTestServer(http.StatusOK, nil)
	defer server.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	msg := &xapp.RMRParams{}

	control := NewControl()
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestQoEPrediction(msg)
}

func TestNegativeHandleRequestQoEPrediction_WhenResponseStatusBadRequest_ExpectReturn(t *testing.T) {
	server := createPostTestServer(http.StatusBadRequest, nil)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Setenv("RIC_MSG_BUF_CHAN_LEN", "256")
	t.Setenv("MLXAPP_HEADERHOST", "qoe-model.kserve-test.example.com")
	t.Setenv("MLXAPP_HOST", strings.Join(strings.Split(server.URL, ":")[:2], ":"))
	t.Setenv("MLXAPP_PORT", strings.Split(server.URL, ":")[2])
	t.Setenv("MLXAPP_REQURL", "v1/models/qoe-model:predict")

	pr, _ := json.Marshal(data.QoePredictionRequest{
		PredictionUE: "12345",
		UEMeasurement: data.UEMeasurementType{
			ServingCellID:            "310-680-200-555002",
			MeasTimestampUEPDCPBytes: "2020-03-18 02:23:18.220",
			MeasPeriodUEPDCPBytes:    20,
			UEPDCPBytesDL:            2500000,
			UEPDCPBytesUL:            1000000,
			MeasTimestampUEPRBUsage:  "2020-03-18 02:23:18.220",
			MeasPeriodUEPRBUsage:     20,
			UEPRBUsageDL:             10,
			UEPRBUsageUL:             30,
		},
		CellMeasurements: []data.CellMeasurement{
			data.CellMeasurement{
				CellID:                 "310-680-220-555001",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            250000,
				PDCPBytesUL:            100000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             30,
				AvailPRBUL:             50,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -90, RSRQ: -13, RSSINR: -2.5},
			},
			data.CellMeasurement{
				CellID:                 "310-680-220-555003",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            200000,
				PDCPBytesUL:            120000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             60,
				AvailPRBUL:             80,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -140, RSRQ: -17, RSSINR: -6},
			},
			data.CellMeasurement{
				CellID:                 "310-680-220-555002",
				MeasTimestampPDCPBytes: "2020-03-18 02:23:18.220",
				MeasPeriodPDCPBytes:    20,
				PDCPBytesDL:            190000,
				PDCPBytesUL:            100000,
				MeasTimestampAvailPRB:  "2020-03-18 02:23:18.220",
				MeasPeriodAvailPRB:     20,
				AvailPRBDL:             30,
				AvailPRBUL:             45,
				MeasTimestampRF:        "2020-03-18 02:23:18.220",
				MeasPeriodRF:           40,
				RFMeasurements:         data.RFMeasurement{RSRP: -115, RSRQ: -16, RSSINR: -5},
			},
		},
	})

	msg := &xapp.RMRParams{
		Payload: pr,
	}

	control := NewControl()
	control.RmrCommand = mocks_control.NewFakeRMRClient()

	control.handleRequestQoEPrediction(msg)
}
