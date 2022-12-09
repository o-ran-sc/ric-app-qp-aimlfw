.. This work is licensed under a Creative Commons Attribution 4.0 International License.
.. http://creativecommons.org/licenses/by/4.0

.. Copyright (c) 2022 Samsung Electronics Co., Ltd. All Rights Reserved.

QoE Prediction assist xApp Overview
===================================

QoE Prediction assist xApp(ric-app-qp-aimlfw) is an xApp that supports QoE Prediction on the AIMLFW, and an xApp of the Traffic Steering O-RAN usecase.
The difference from the existing QoE Prediction(ric-app-qp) is that the function to interact with the MLxApp of AIMLFW is added and the inference function is removing.
The main operations are as follows:

#. Traffic Steering xApp transmits prediction request to QoE Prediction assist xApp.
#. QoE Prediction assist xApp builds prediction request message using cell metrics from influxdb and then sends prediction request to MLxApp. Cell Metrics are stored in the influxDB by the KPIMON xApp.
#. QoE Prediction assist xApp receives the result of prediction from MLxApp.
#. QoE Prediction assist xApp transmits the received prediction result to Traffic Sterring xApp.


Expected Input
--------------
QoE Prediction assist xApp expects the following message along with the `TS_UE_LIST` message type through RMR.

 .. code:: none

{"UEPredictionSet": ["Car-1"]}


Expected Output
---------------
QoE Prediction assist xApp transmits the following message along with the `TS_QOE_PREDICTION` message type throgh RMR.
The message below is the prediction result for both downlink and uplink throughput.

.. code:: none

 {"Car-1":{
 "c6/B2": [12650, 12721],
 "c6/N77": [12663, 12739],
 "c1/B13": [12576, 12655],
 "c7/B13": [12649, 12697],
 "c5/B13": [12592, 12688]
 }}