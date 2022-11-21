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

package data

type Timestamp struct {
	TVsec  int64 `json:"tv_sec"`
	TVnsec int64 `json:"tv_nsec"`
}

type CellMetricsEntry struct {
	MeasTimestampPDCPBytes Timestamp `json:"MeasTimestampPDCPBytes"`
	CellID                 string    `json:"CellID"`
	PDCPBytesDL            int64     `json:"PDCPBytesDL"`
	PDCPBytesUL            int64     `json:"PDCPBytesUL"`
	MeasTimestampPRB       Timestamp `json:"MeasTimestampAvailPRB"`
	AvailPRBDL             int64     `json:"AvailPRBDL"`
	AvailPRBUL             int64     `json:"AvailPRBUL"`
	MeasPeriodPDCP         int64     `json:"MeasPeriodPDCPBytes"`
	MeasPeriodPRB          int64     `json:"MeasPeriodAvailPRB"`
}

type QoePredictionInput struct {
	SignatureName string        `json:"signature_name"`
	Instances     [][][]float32 `json:"instances"`
}

type PredictRequest struct {
	UEPredictionSet []string `json:"UEPredictionSet"` // {"UEPredictionSet": ["Car-1"]}
}
