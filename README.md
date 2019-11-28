# thunderball
Converts Prometheus alertmanager firing alerts to a jira issue.

## Default Alertmanager request

A typical Prometheus alertmanager json payload:

```
{
  "receiver": "jira-notify",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "InstanceDown",
        "instance": "172.17.0.8:9200",
        "job": "javaapp",
        "severity": "critical"
      },
      "annotations": {
        "description": "172.17.0.8:9200 of job *javaapp* has been down for more than 1 mi",
        "summary": "Instance 172.17.0.8:9200 down"
      },
      "startsAt": "2019-10-31T15:15:25.84136864Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://bf6dfb8d75a3:9090/graph?g0.expr=up+%3D%3D+0&g0.tab=1",
      "fingerprint": "50a7edff2fb8efff"
    }
  ],
  "groupLabels": {
    "instance": "172.17.0.8:9200",
    "severity": "critical"
  },
  "commonLabels": {
    "alertname": "InstanceDown",
    "instance": "172.17.0.8:9200",
    "job": "javaapp",
    "severity": "critical"
  },
  "commonAnnotations": {
    "description": "172.17.0.8:9200 of job *javaapp* has been down for more than 1 mi",
    "summary": "Instance 172.17.0.8:9200 down"
  },
  "externalURL": "http://b22de154018f:9093",
  "version": "4",
  "groupKey": "{}/{alertname=\"InstanceDown\"}:{instance=\"172.17.0.8:9200\", severity=\"critical\"}"
}
```
