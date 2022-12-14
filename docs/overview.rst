.. This work is licensed under a Creative Commons Attribution 4.0 International License.
.. http://creativecommons.org/licenses/by/4.0

.. Copyright (c) 2022 Samsung Electronics Co., Ltd. All Rights Reserved.

QoE Prediction assist xApp Overview
===================================

QoE Prediction assist xApp(ric-app-qp-aimlfw) is an xApp that supports QoE Prediction on the AIMLFW, and an xApp of the Traffic Steering O-RAN usecase.
The difference from the existing QoE Prediction(ric-app-qp) is that the function to interact with the MLxApp of AIMLFW is added and the inference function is removing.
The main operations are as follows:

#. QP Driver xApp transmits prediction request to QoE Prediction assist xApp.
#. QoE Prediction assist xApp builds prediction request message and then sends prediction request to MLxApp.
#. QoE Prediction assist xApp receives the result of prediction from MLxApp.
#. QoE Prediction assist xApp transmits the received prediction result to Traffic Sterring xApp.


Expected Input
--------------
QoE Prediction assist xApp expects the following message along with the `TS_QOE_PRED_REQ` message type through RMR.

.. code-block:: none 

    {
        PredictionUE: "12345",
		UEMeasurement: {
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
		CellMeasurements: [
			{
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
			{
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
			{
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
			}]
	}


Expected Output
---------------
QoE Prediction assist xApp transmits the following message along with the `TS_QOE_PREDICTION` message type throgh RMR.
The message below is the prediction result for both downlink and uplink throughput.

.. code-block:: none 

 {"Car-1":{
 "c6/B2": [12650, 12721],
 "c6/N77": [12663, 12739],
 "c1/B13": [12576, 12655],
 "c7/B13": [12649, 12697],
 "c5/B13": [12592, 12688]
 }}