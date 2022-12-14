.. This work is licensed under a Creative Commons Attribution 4.0 International License.
.. http://creativecommons.org/licenses/by/4.0

.. Copyright (c) 2022 Samsung Electronics Co., Ltd. All Rights Reserved.

User-Guide
================

.. contents::
   :depth: 3
   :local:

Overview
--------
- The entire operation scenario can be checked when deployed with KPIMON xApp, Traffic Steering xApp and QP Driver xApp.
- This is compatible with influxDB2.

Build Image
-----------
Use the `docker build` command for docker image build.

.. code-block:: none 

    qp-aimlfw $ docker build -f Dockerfile .

    Setting up curl (7.68.0-1ubuntu2.14) ...
    Processing triggers for libc-bin (2.31-0ubuntu9.9) ...
    Processing triggers for ca-certificates (20211016ubuntu0.20.04.1) ...
    Updating certificates in /etc/ssl/certs...
    0 added, 0 removed; done.
    Running hooks in /etc/ca-certificates/update.d...
    done.
    Removing intermediate container 694003ecb2e1
     ---> 18d14cdf7365
    Step 15/19 : COPY --from=builder /opt/qoe-aiml-assist/build/qoe-aiml-assist .
      ---> 20d6be2c65b1
    Step 16/19 : COPY --from=builder /usr/local/include /usr/local/include
      ---> e5f0c6465df4
    Step 17/19 : COPY --from=builder /usr/local/lib /usr/local/lib
      ---> 5bcaca789e8c
    Step 18/19 : COPY --from=builder /opt/qoe-aiml-assist/config/* /opt/ric/config/
      ---> a5ef330d90df
    Step 19/19 : RUN ldconfig
      ---> Running in b6ee4f81f4ff
    Removing intermediate container b6ee4f81f4ff
      ---> 91d0473a7cf1
    Successfully built 91d0473a7cf1
    Successfully tagged nexus3.o-ran-sc.org:10004/o-ran-sc/ric-app-qp-aimlfw-docker:1.0.0


Environments of qp-aimlfw
-------------------------
By modifying the `config/config-file.json` file before onboarding, we can change the environment variable of the QoE Prediction assist xApp.
Below is a description of the configurable environmental variables related to QoE xApp.

+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_URL`            | URL contianing the address and port of influxDB. ex) "http://127.0.0.1:8086"   |
+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_TOKEN`          |a influx token composed of alphabets and numbers.                               |
+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_BUCKET`         | a bucket name of influxDB.                                                     |
+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_ORG`            | organization name of influxDB.                                                 |
+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_QUERY_START`    | start point of the influxDB query range ex) "-3m" or "-10s"                    |
+------------------------+--------------------------------------------------------------------------------+
|`INFLUX_QUERY_STOP`     | end point of the influxDB query range. If this value is not set like `""`,     |
|                        | the end point is set to the current time(now). ex) "-1m" or ""                 |
+------------------------+--------------------------------------------------------------------------------+
|`MLXAPP_REQ_HEADER_HOST`| header of MLxApp's host. ex) "qoe-model.kserve-test.example.com"               |
+------------------------+--------------------------------------------------------------------------------+
|`MLXAPP_HOST`           | host address of MLxApp. ex) "http://127.0.0.1"                                 |
+------------------------+--------------------------------------------------------------------------------+
|`MLXAPP_PORT`           | port number of MLxApp. ex) "2222"                                              |
+------------------------+--------------------------------------------------------------------------------+
|`MLXAPP_REQ_URL`        |requet URL of MLxapp. ex) "v1/models/qoe-model:predict"                         |
+------------------------+--------------------------------------------------------------------------------+


Onboarding of qp-aimlfw using `dms_cli` tool
---------------------------------------------
Before deploying QoE Prediction assist xApp, onboarding should be done using `config-file.json` and `schema.json`
Onboarding is required for the first time before deploy, and must be performed whenever the value of `config-file.json` is changed.
For onboarding, `dms_cli` and helm3 are required. `dms_cli` can be referred to in this `documentation <https://docs.o-ran-sc.org/projects/o-ran-sc-it-dep/en/latest/installation-guides.html#ric-applications>`

.. code-block:: none 

    $ dms_cli onboard --config_file_path=config/config-file.json --schema_file_path=config/schema.json

    httpGet:
    path: '{{ index .Values "readinessProbe" "httpGet" "path" | toJson }}'
    port: '{{ index .Values "readinessProbe" "httpGet" "port" | toJson }}'
    initialDelaySeconds: '{{ index .Values "readinessProbe" "initialDelaySeconds" | toJson }}'
    periodSeconds: '{{ index .Values "readinessProbe" "periodSeconds" | toJson }}'

    httpGet:
    path: '{{ index .Values "livenessProbe" "httpGet" "path" | toJson }}'
    port: '{{ index .Values "livenessProbe" "httpGet" "port" | toJson }}'
    initialDelaySeconds: '{{ index .Values "livenessProbe" "initialDelaySeconds" | toJson }}'
    periodSeconds: '{{ index .Values "livenessProbe" "periodSeconds" | toJson }}'

    {
    "status": "Created"
    }


Deployment of qp-aimlfw using `dms_cli` tool
---------------------------------------------
Deploy the onboarded QoE Prediction assist xApp using `dms_cli`.

.. code-block:: none 

    $ dms_cli install --xapp_chart_name=qoe-aiml-assist --version=1.0.0 --namespace=ricxapp

    status: OK

    Check if QoE Prediction assist xApp deployed normally.

.. code-block:: none 

    $ kubectl get pods -n ricxapp
    NAME                                       READY   STATUS    RESTARTS   AGE
    ricxapp-qoe-aiml-assist-5f788bb667-47k5h   0/1     Pending   0          3m


    $ kubectl get svc -n=ricxapp
    NAME                                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
    aux-entry                              ClusterIP   10.106.133.25   <none>        80/TCP,443/TCP      8d
    service-ricxapp-qoe-aiml-assist-http   ClusterIP   10.96.95.160    <none>        8080/TCP            129m
    service-ricxapp-qoe-aiml-assist-rmr    ClusterIP   10.107.182.86   <none>        4560/TCP,4561/TCP   129m
