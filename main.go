package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type config struct {
	Port          int    `env:"PORT" envDefault:"7337"`
	AlertEnv      string `env:"ALERT_ENV" envDefault:"Dev"`
	AlertPriority string `env:"ALERT_PRIORITY" envDefault:"High"`
	// Default component
	AlertComponent string `env:"ALERT_COMPONENT" envDefault:"platform"`
	// Project Key name
	DefaultProject string `env:"DEFAULT_PROJECT" envDefault:"ABC"`
	// Epic Link
	AlertEpicLink string `env:"ALERT_EPICLINK" envDefault:"AM-1"`
	// Jira URL https://<project>.atlassian.ent
	JiraURL              string `env:"JIRA_URL" envDefault:"https://project.atlassian.net"`
	JiraUser             string `env:"JIRA_USER,required"`
	JiraApiToken         string `env:"JIRA_API_TOKEN,required"`
	JiraIssueTemplateURL string `env:"JIRA_TEMPLATE_URL"`
}

// Note that if there are custom fields changes, those needs to be reflected here.
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

// PrometheusAlert struct to handle with prometheus alertmanager payload (v4).
type prometheusAlert struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Version  string `json:"version"`
	Alerts   []struct {
		Status string `json:"status"`
		Labels struct {
			AlertName string `json:"alertname"`
			Instance  string `json:"instance"`
			Job       string `json:"job"`
			Severity  string `json:"severity"`
		}
		Annotations struct {
			Description string `json:"description"`
			Summary     string `json:"summary"`
		}
	}
}

// Make standard loggin available globally
var ginLog io.Writer

// jsonTemplate This function returns the json template based on the JIRA_TEMPLATE_URL env variable.
// If the variable is set then it fetches the json template following the http endpoint on the variable.
// If not set then uses the hardcoded template. const jiraJsonTemplate
func jsonTemplate(cfg config) string {
	templateResult := jiraJsonTemplate

	if cfg.JiraIssueTemplateURL != "" {
		resp, err := http.Get(cfg.JiraIssueTemplateURL)
		if err != nil {
			ginLog.Write([]byte(err.Error()))
			return templateResult
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		templateResult = buf.String()
	}

	return templateResult
}

//setupRouter Sets all the URI routing available for the service.
func setupRouter(cfg config, jsonTemplate string) *gin.Engine {
	r := gin.Default()
	// Enable go Prometheus exporter metrics - More to add
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// https://community.atlassian.com/t5/Jira-questions/Status-URL-for-monitoring/qaq-p/146530
	r.GET("/healthz", func(c *gin.Context) {
		resp, err := http.Get(cfg.JiraURL + "/status")
		if err != nil {
			ginLog.Write([]byte(err.Error()))
		}
		ginLog.Write([]byte(fmt.Sprintf("GET /status: Upstream response: %s\n", resp)))

		defer resp.Body.Close()

		c.JSON(http.StatusOK, gin.H{"status": "healthy", "upstream": map[bool]string{true: "available", false: "unavailable"}[resp.StatusCode == 200]})
	})

	r.POST("/jira", func(c *gin.Context) {

		var payload prometheusAlert

		if c.BindJSON(&payload) == nil {

			var b bytes.Buffer
			t, err := template.New("JiraTemplate").Parse(jsonTemplate)
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}

			err = t.Execute(&b, struct {
				Summary, Description, EpicLink, Project, Component, Environment, Priority string
			}{
				Summary:     fmt.Sprintf("%s: %s %s (%s)", payload.Alerts[0].Labels.Severity, payload.Alerts[0].Labels.AlertName, payload.Alerts[0].Labels.Job, payload.Alerts[0].Labels.Instance),
				Description: payload.Alerts[0].Annotations.Description,
				EpicLink:    cfg.AlertEpicLink,
				Project:     cfg.DefaultProject,
				Component:   cfg.AlertComponent,
				Environment: cfg.AlertEnv,
				Priority:    cfg.AlertPriority,
			})
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}

			client := &http.Client{}
			req, err := http.NewRequest("POST", cfg.JiraURL+"/rest/api/2/issue", bytes.NewReader(b.Bytes()))
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}

			// All these are required for a BasicAuth with Jira.
			req.SetBasicAuth(cfg.JiraUser, cfg.JiraApiToken) // Performs Base64 hash.
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "*/*")
			// if required to remove gzip encoding add the following:
			// req.Header.Set("Accept-Encoding", "identity")

			resp, err := client.Do(req)
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}
			ginLog.Write([]byte(fmt.Sprintf("POST /jira: Upstream response: %s\n", resp)))

			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				c.JSON(resp.StatusCode, gin.H{"status": "upstream server error", "upstream_response": resp.StatusCode})
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}
			rJSON := make(map[string]interface{})
			err = json.Unmarshal(body, &rJSON)
			if err != nil {
				ginLog.Write([]byte(err.Error()))
			}

			req.Body.Close()
			if resp.StatusCode == 201 {
				c.JSON(http.StatusOK, gin.H{"status": "ok",
					"issue": rJSON["key"],
					"url":   rJSON["self"]})
			}
		}
	})
	return r
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	ginLog = gin.DefaultErrorWriter

	r := setupRouter(cfg, jsonTemplate(cfg))

	// Listen and Serve at 0.0.0.0:cfg.Port
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
