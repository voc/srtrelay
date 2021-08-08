# srtrelay API
See [config.toml.example](../config.toml.example) for configuring the API endpoint.

## Stream status - /streams
- Returns a list of active streams with additional statistics.
- Content-Type: application/json
- Example:
```
GET http://localhost:8080/streams

[{"name":"abc","clients":0,"created":"2020-11-24T23:55:27.265206348+01:00"}]
```

## Socket statistics - /sockets
- Returns internal srt statistics for each SRT client
  - the exact statistics might change depending over time
  - this will show stats for both publishers and subscribers
- Content-Type: application/json
- Example:
```json
[
  {
    "address": "127.0.0.1:59565",
    "stream_id": "publish/q2",
    "stats": {
      "MsTimeStamp": 26686,
      "PktSentTotal": 0,
      "PktRecvTotal": 9484,
      "PktSndLossTotal": 0,
      "PktRcvLossTotal": 0,
      "PktRetransTotal": 0,
      "PktSentACKTotal": 1680,
      "PktRecvACKTotal": 0,
      "PktSentNAKTotal": 0,
      "PktRecvNAKTotal": 0,
      "UsSndDurationTotal": 0,
      "PktSndDropTotal": 0,
      "PktRcvDropTotal": 0,
      "PktRcvUndecryptTotal": 0,
      "ByteSentTotal": 0,
      "ByteRecvTotal": 11866496,
      "ByteRcvLossTotal": 0,
      "ByteRetransTotal": 0,
      "ByteSndDropTotal": 0,
      "ByteRcvDropTotal": 0,
      "ByteRcvUndecryptTotal": 0,
      "PktSent": 0,
      "PktRecv": 9484,
      "PktSndLoss": 0,
      "PktRcvLoss": 0,
      "PktRetrans": 0,
      "PktRcvRetrans": 0,
      "PktSentACK": 1680,
      "PktRecvACK": 0,
      "PktSentNAK": 0,
      "PktRecvNAK": 0,
      "MbpsSendRate": 0,
      "MbpsRecvRate": 3.557279995149639,
      "UsSndDuration": 0,
      "PktReorderDistance": 0,
      "PktRcvAvgBelatedTime": 0,
      "PktRcvBelated": 0,
      "PktSndDrop": 0,
      "PktRcvDrop": 0,
      "PktRcvUndecrypt": 0,
      "ByteSent": 0,
      "ByteRecv": 11866496,
      "ByteRcvLoss": 0,
      "ByteRetrans": 0,
      "ByteSndDrop": 0,
      "ByteRcvDrop": 0,
      "ByteRcvUndecrypt": 0,
      "UsPktSndPeriod": 10,
      "PktFlowWindow": 8192,
      "PktCongestionWindow": 8192,
      "PktFlightSize": 0,
      "MsRTT": 0.013,
      "MbpsBandwidth": 1843.584,
      "ByteAvailSndBuf": 12288000,
      "ByteAvailRcvBuf": 12160500,
      "MbpsMaxBW": 1000,
      "ByteMSS": 1500,
      "PktSndBuf": 0,
      "ByteSndBuf": 0,
      "MsSndBuf": 0,
      "MsSndTsbPdDelay": 200,
      "PktRcvBuf": 74,
      "ByteRcvBuf": 89851,
      "MsRcvBuf": 188,
      "MsRcvTsbPdDelay": 200,
      "PktSndFilterExtraTotal": 0,
      "PktRcvFilterExtraTotal": 0,
      "PktRcvFilterSupplyTotal": 0,
      "PktRcvFilterLossTotal": 0,
      "PktSndFilterExtra": 0,
      "PktRcvFilterExtra": 0,
      "PktRcvFilterSupply": 0,
      "PktRcvFilterLoss": 0,
      "PktReorderTolerance": 0
    }
  }
]
```