# thunderball
A service that converts Prometheus alertmanager firing alerts to jira issues.

## Details
Thunderball converts Prometheus alertmanager alerts to jira issues. Majority of the fields are configureable via environment variables. Thunderball contains a default jira issue template (jiraJsonTemplate) but it can read templates from an external http/https endpoints. 

## Default Jira template
```
const jiraJsonTemplate = `{
    "fields": {
       "customfield_10008": "{{ .EpicLink}}",
       "project":
       {
          "key": "{{ .Project}}"
       },
       "summary": "{{ .Summary}}",
       "description": "{{ .Description}}",
       "issuetype": {
          "name": "Bug"
       },
       "customfield_10019": "none",
       "customfield_10020": "none",
       "customfield_10021": "none",
       "customfield_10022": [
		{ "self": "https://project.atlassian.net/rest/api/2/customFieldOption/10007",
		  "value" : "{{ .Environment}}"
		}
	   ],
       "components": [
		{ "name": "{{ .Component}}"}
		],
	   "priority":
		{
			"name": "{{ .Priority}}"
		} 
	   }
    }`
 ```
 
## How to build thunderball
```
make build
go mod download
go build -o thunderball .
```
## Running
### Docker
docker run --name thunderball -d -p 7337:7337 -e JIRA_USER="user@mydomain.com" -e JIRA_API_TOKEN="apitoken" thunderball
### CMD
JIRA_USER="user@mydomain.com" JIRA_API_TOKEN="apitoken" ./thunderball 

## How to plug into alertmanager
Add the following receiver (and thunderball ipaddress) to alertmanager configuration
```
- name: jira-notify
  webhook_configs:
  - url: "http://${thunderball_ip}:7337/jira"
    send_resolved: false
```

## (for guidance) Default Alertmanager request
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
