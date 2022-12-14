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
	"fmt"
	"net/http"
	"os"
	"strconv"

	"gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/data"
	"gerrit.o-ran-sc.org/r/ric-app/qp-aimlfw/influx"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

const (
	DEFAULT_MSG_BUF_CHAN_LEN int = 256
	SIGNITURE_NAME               = "serving_default"

	ENV_RIC_MSG_BUF_CHAN_LEN = "RIC_MSG_BUF_CHAN_LEN"
	ENV_INFLUX_URL           = "INFLUX_URL"
	ENV_WAIT_SDL             = "db.waitForSdl"

	ENV_MLXAPP_PREFIX          = "MLXAPP"
	ENV_MLXAPP_REQ_HEADER_HOST = "HEADERHOST"
	ENV_MLXAPP_HOST            = "HOST"
	ENV_MLXAPP_PORT            = "PORT"
	ENV_MLXAPP_REQ_URL         = "REQURL"
)

type RmrCommand interface {
	Send(*xapp.RMRParams, bool) bool
	SendRts(*xapp.RMRParams) bool
	GetRicMessageName(int) string
}

type RmrWrapper struct {
}

func (rw RmrWrapper) Send(msg *xapp.RMRParams, isRts bool) bool {
	return xapp.Rmr.Send(msg, isRts)
}

func (rw RmrWrapper) SendRts(msg *xapp.RMRParams) bool {
	return xapp.Rmr.SendRts(msg)
}

func (rw RmrWrapper) GetRicMessageName(id int) string {
	return xapp.Rmr.GetRicMessageName(id)
}

type Control struct {
	RmrCommand
	influxDB      influx.InfluxDBCommand
	rcChan        chan *xapp.RMRParams
	mlxAppConfigs MLxAppConfigs
}

type MLxAppConfigs struct {
	HeaderHost string
	Host       string
	Port       string
	ReqUrl     string
}

func init() {
	viper.SetEnvPrefix(ENV_MLXAPP_PREFIX)
	viper.BindEnv(ENV_MLXAPP_REQ_HEADER_HOST)
	viper.BindEnv(ENV_MLXAPP_HOST)
	viper.BindEnv(ENV_MLXAPP_PORT)
	viper.BindEnv(ENV_MLXAPP_REQ_URL)
}

func NewControl() Control {
	influxDB := influx.CreateInfluxDB()
	ricMsgBufChanLen, _ := getEnvAndSetInt(DEFAULT_MSG_BUF_CHAN_LEN, ENV_RIC_MSG_BUF_CHAN_LEN)

	var mlxAppConfigs MLxAppConfigs
	err := viper.Unmarshal(&mlxAppConfigs)
	if err != nil {
		xapp.Logger.Error("failed to Unmarshal MLxAppConfigs")
	}
	out, err := json.Marshal(mlxAppConfigs)
	if err != nil {
		xapp.Logger.Error("failed to json.Marshal MLxAppConfigs")
	}
	xapp.Logger.Debug("MLxAppConfigs : %s", out)

	return Control{
		influxDB:      &influxDB,
		rcChan:        make(chan *xapp.RMRParams, ricMsgBufChanLen),
		mlxAppConfigs: mlxAppConfigs,
		RmrCommand:    RmrWrapper{},
	}
}

func getEnvAndSetInt(val int, envKey string) (int, bool) {
	envStr, envFlag := os.LookupEnv(envKey)

	if !envFlag {
		xapp.Logger.Error("failed to read %s from env, use default %s(%d)", envKey, envKey, val)
		return val, envFlag
	}

	val, _ = strconv.Atoi(envStr)
	xapp.Logger.Info("read to %s from env, %s(%d)", envKey, envKey, val)

	return val, envFlag
}

func (c *Control) xAppStartCB(d interface{}) {
	go c.controlLoop()
}

func (c *Control) Run() {
	xapp.Logger.SetMdc("qoe-aiml-assist", "1.0.0")
	xapp.SetReadyCB(c.xAppStartCB, true)
	waitForSdl := xapp.Config.GetBool(ENV_WAIT_SDL)
	xapp.RunWithParams(c, waitForSdl)
}

func (c *Control) sendRequestToMLxApp(jsonMsg []byte) []byte {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Host", c.mlxAppConfigs.HeaderHost).
		EnableTrace().
		SetBody(jsonMsg).
		Post(fmt.Sprintf("%s:%s/%s", c.mlxAppConfigs.Host, c.mlxAppConfigs.Port, c.mlxAppConfigs.ReqUrl))

	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		xapp.Logger.Error("failed to POST : err = %s, resp = %s, code = %d", err, resp, resp.StatusCode())
		return nil
	}

	xapp.Logger.Info("Response from MLxApp : %s", resp)
	return resp.Body()
}

func (c *Control) handleRequestQoEPrediction(msg *xapp.RMRParams) {
	var qoePredictionRequest data.QoePredictionRequest
	err := json.Unmarshal(msg.Payload, &qoePredictionRequest)
	if err != nil {
		xapp.Logger.Error("filed to unmarshal msg : %s", err)
		return
	}
	xapp.Logger.Info("Received Msg : %+v", qoePredictionRequest)

	var qoePredictionInput data.QoePredictionInput
	qoePredictionInput.SignatureName = SIGNITURE_NAME

	for i := 0; i < len(qoePredictionRequest.CellMeasurements); i++ {
		qoePredictionInput.Instances = append(qoePredictionInput.Instances, [][]float32{})
		qoePredictionInput.Instances[i] = append(qoePredictionInput.Instances[i], []float32{float32(qoePredictionRequest.CellMeasurements[i].PDCPBytesUL), float32(qoePredictionRequest.CellMeasurements[i].PDCPBytesDL)})
	}
	xapp.Logger.Info("Qoe Prediction Request = %v", qoePredictionInput)

	jsonbytes, err := json.Marshal(qoePredictionInput)
	if err != nil {
		xapp.Logger.Error("fail to marshal : %s", err)
		return
	}

	response := c.sendRequestToMLxApp(jsonbytes)
	if response == nil {
		xapp.Logger.Error("fail to request prediction to MLxApp")
		return
	}

	var qoePredictionResult data.QoePredictionResult
	err = json.Unmarshal(response, &qoePredictionResult)
	if err != nil {
		xapp.Logger.Error("filed to unmarshal msg : %s", err)
		return
	}
	xapp.Logger.Debug("Unmarshaled response: %+v", qoePredictionResult)

	predictions := make(map[string][]float32)
	for i := 0; i < len(qoePredictionRequest.CellMeasurements); i++ {
		predictions[qoePredictionRequest.CellMeasurements[i].CellID] = []float32{qoePredictionResult.Predictions[i][0], qoePredictionResult.Predictions[i][1]}
	}
	predResultMsg := make(map[string]map[string][]float32)
	predResultMsg[qoePredictionRequest.PredictionUE] = predictions

	jsonBytes, err := json.Marshal(predResultMsg)
	if err != nil {
		xapp.Logger.Error("filed to marshal msg : %s", err)
		return
	}
	xapp.Logger.Info("QoE Prediction Result Msg : " + string(jsonBytes))

	c.sendPredictionResult(msg, jsonBytes)
}

func (c *Control) handleRequestPrediction(msg *xapp.RMRParams) {
	var predictRequest data.PredictRequest
	err := json.Unmarshal(msg.Payload, &predictRequest)
	if err != nil {
		xapp.Logger.Error("failed to unmarshal msg : %s", err)
		return
	}

	ueid := predictRequest.UEPredictionSet[0]
	xapp.Logger.Info("requested UEPredictionSet = %s", ueid)

	cellMetricsEntries, err := c.influxDB.RetrieveCellMetrics()
	if err != nil {
		xapp.Logger.Error("failed to RetrieveCellMetrics")
		return
	}

	if cellMetricsEntries == nil || len(cellMetricsEntries) == 0 {
		xapp.Logger.Error("CellMetrics is null !")
		return
	}

	qoePrectionInput := c.makeRequestPredictionMsg(cellMetricsEntries)
	jsonbytes, err := json.Marshal(qoePrectionInput)
	if err != nil {
		xapp.Logger.Error("fail to marshal : %s", err)
		return
	}

	response := c.sendRequestToMLxApp(jsonbytes)
	if response == nil {
		xapp.Logger.Error("fail to request prediction to MLxApp")
		return
	}

	c.sendPredictionResult(msg, response)
}

func (c *Control) makeRequestPredictionMsg(cellMetricsEntries []data.CellMetricsEntry) data.QoePredictionInput {
	var qoePredictionInput data.QoePredictionInput
	qoePredictionInput.SignatureName = SIGNITURE_NAME

	for i := 0; i < len(cellMetricsEntries); i++ {
		qoePredictionInput.Instances = append(qoePredictionInput.Instances, [][]float32{})
		qoePredictionInput.Instances[i] = append(qoePredictionInput.Instances[i], []float32{float32(cellMetricsEntries[i].PDCPBytesUL), float32(cellMetricsEntries[i].PDCPBytesDL)})
	}
	return qoePredictionInput
}

func (c *Control) controlLoop() {
	for {
		msg := <-c.rcChan
		xapp.Logger.Debug("Received message type: %d", msg.Mtype)
		switch msg.Mtype {
		case xapp.TS_UE_LIST:
			go c.handleRequestPrediction(msg)
		case xapp.TS_QOE_PRED_REQ:
			go c.handleRequestQoEPrediction(msg)
		default:
			xapp.Logger.Info("Unknown message type '%d', discarding", msg.Mtype)
		}
	}
}

func (c *Control) sendPredictionResult(msg *xapp.RMRParams, respBody []byte) {
	msg.Mtype = xapp.TS_QOE_PREDICTION
	msg.PayloadLen = len(respBody)
	msg.Payload = respBody
	ret := c.Send(msg, false)
	xapp.Logger.Info("result of SendPredictionResult = %t", ret)
}

func (c *Control) Consume(msg *xapp.RMRParams) (err error) {
	id := c.GetRicMessageName(msg.Mtype)
	xapp.Logger.Info(
		"Message received: name=%s meid=%s subId=%d txid=%s len=%d",
		id,
		msg.Meid.RanName,
		msg.SubId,
		msg.Xid,
		msg.PayloadLen,
	)
	c.rcChan <- msg
	return nil
}
