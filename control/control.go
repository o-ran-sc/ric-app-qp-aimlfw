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

	"gerrit.o-ran-sc.org/r/qp-aiml/data"
	"gerrit.o-ran-sc.org/r/qp-aiml/influx"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/go-resty/resty/v2"
)

const (
	DEFAULT_MSG_BUF_CHAN_LEN int = 256
	SIGNITURE_NAME               = "serving_default"

	ENV_RIC_MSG_BUF_CHAN_LEN   = "ricMsgBufChanLen"
	ENV_INFLUX_URL             = "INFLUX_URL"
	ENV_WAIT_SDL               = "db.waitForSdl"
	ENV_MLXAPP_REQ_HEADER_HOST = "MLXAPP_REQ_HEADER_HOST"
	ENV_MLXAPP_HOST            = "MLXAPP_HOST"
	ENV_MLXAPP_PORT            = "MLXAPP_PORT"
	ENV_MLXAPP_REQ_URL         = "MLXAPP_REQ_URL"
)

type Control struct {
	influxClient influx.InfluxClient
	rcChan       chan *xapp.RMRParams
}

func NewControl() Control {
	influxClient := influx.CreateInfluxClient()
	ricMsgBufChanLen, _ := getEnvAndSetInt(DEFAULT_MSG_BUF_CHAN_LEN, ENV_RIC_MSG_BUF_CHAN_LEN)
	return Control{influxClient, make(chan *xapp.RMRParams, ricMsgBufChanLen)}
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

func (c *Control) handleRequestPrediction(ranName string, msg *xapp.RMRParams) {

	var predictRequest data.PredictRequest
	err := json.Unmarshal(msg.Payload, &predictRequest)
	if err != nil {
		xapp.Logger.Error("failed to unmarshal msg : %s", err)
		return
	}
	ueid := predictRequest.UEPredictionSet[0]
	xapp.Logger.Info("requested UEPredictionSet = %s", ueid)

	cellMetricsEntries, err := c.influxClient.RetrieveCellMetrics()
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

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Host", os.Getenv(ENV_MLXAPP_REQ_HEADER_HOST)).
		EnableTrace().
		SetBody(jsonbytes).
		Post(fmt.Sprintf("%s:%s/%s", os.Getenv(ENV_MLXAPP_HOST), os.Getenv(ENV_MLXAPP_PORT), os.Getenv(ENV_MLXAPP_REQ_URL)))

	if err != nil || resp == nil || resp.StatusCode() != http.StatusOK {
		xapp.Logger.Error("failed to POST : err = %s, resp = %s, code = %s, sendmsg = %s", err, resp, resp.StatusCode(), qoePrectionInput)
		return
	}

	xapp.Logger.Info("Response from MLxApp : %s", resp)
	c.sendPredictionResult(msg, resp.Body())
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
			go c.handleRequestPrediction(msg.Meid.RanName, msg)
		default:
			xapp.Logger.Info("Unknown message type '%d', discarding", msg.Mtype)
		}
	}
}

func (c *Control) sendPredictionResult(msg *xapp.RMRParams, respBody []byte) {
	msg.Mtype = xapp.TS_QOE_PREDICTION
	msg.PayloadLen = len(respBody)
	msg.Payload = respBody
	ret := xapp.Rmr.SendRts(msg)
	xapp.Logger.Info("result of SendPredictionResult = %s", ret)
}

func (c *Control) Consume(msg *xapp.RMRParams) (err error) {
	id := xapp.Rmr.GetRicMessageName(msg.Mtype)
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
