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

func createPostTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			if req.URL.Path == "/v1/models/qoe-model:predict" {
				rw.Header().Set("Content-Type", "application/x-www-form-urlencoded")
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
	server := createPostTestServer()
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
	server := createPostTestServer()
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
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			if req.URL.Path == "/v1/models/qoe-model:predict" {
				rw.Header().Set("Content-Type", "application/json; charset=utf-8")
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}))
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
