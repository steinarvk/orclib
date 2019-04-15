package orcclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// func (c *orcClient) CloseIdleConnections() {
// 	c.httpClient.CloseIdleConnections()
// }

func (c *orcClient) Do(req *http.Request) (*http.Response, error) {
	t0 := time.Now()

	targetCanonicalHost, err := c.checkRequest(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"client": c.clientName,
			"error":  err,
		}).Errorf("Rejected outgoing request")
		metricRequestsRejected.With(prometheus.Labels{
			"client": c.clientName,
			"reason": checkFailedStatsReason(err),
		}).Inc()
		return nil, err
	}

	if err := c.m.addAuth(req, targetCanonicalHost); err != nil {
		logrus.WithFields(logrus.Fields{
			"client": c.clientName,
			"error":  err,
		}).Errorf("Rejected outgoing request: failed to add outer-layer auth")
		metricRequestsRejected.With(prometheus.Labels{
			"client": c.clientName,
			"reason": "failed-to-add-auth",
		}).Inc()
		return nil, fmt.Errorf("Failed to add auth: %v", err)
	}

	labels := prometheus.Labels{
		"client": c.clientName,
		"method": req.Method,
		"target": targetCanonicalHost,
	}
	metricRequestsBegun.With(labels).Inc()

	labels["ok"] = "false"
	labels["code"] = ""

	var finalResponse *http.Response
	var finalError error

	defer func() {
		labels["ok"] = fmt.Sprintf("%v", err == nil)
		if finalResponse != nil {
			labels["code"] = fmt.Sprintf("%d", finalResponse.StatusCode)
		}
		duration := time.Since(t0)

		logrus.WithFields(logrus.Fields{
			"client":   c.clientName,
			"error":    err,
			"target":   targetCanonicalHost,
			"scheme":   req.URL.Scheme,
			"path":     req.URL.Path,
			"duration": duration,
		}).Infof("Finished outgoing request")

		metricRequestsFinished.With(labels).Inc()
		metricRequestsFinishedLatency.With(labels).Observe(duration.Seconds())
	}()

	finalResponse, finalError = c.httpClient.Do(req)
	return finalResponse, finalError
}

type badOrcRequest struct {
	statsReason string
}

func (e badOrcRequest) Error() string {
	return fmt.Sprintf("Bad outgoing request: %v", e.statsReason)
}

func checkFailedStatsReason(err error) string {
	if cast, ok := err.(badOrcRequest); ok {
		return cast.statsReason
	}
	if err == nil {
		return "success"
	}
	return "unknown"
}

func (c *orcClient) checkRequest(req *http.Request) (string, error) {
	if req == nil {
		return "", badOrcRequest{"nil"}
	}
	if req.Method == "" {
		return "", badOrcRequest{"empty-method"}
	}
	if req.URL == nil {
		return "", badOrcRequest{"nil-url"}
	}
	targetHost := req.URL.Host
	if req.URL.Scheme != "https" {
		return "", badOrcRequest{"not-https"}
	}
	if req.URL.User != nil {
		return "", badOrcRequest{"extra-userinfo"}
	}
	if !c.m.isHostAllowed(targetHost) {
		return "", badOrcRequest{"host-not-allowed"}
	}
	return targetHost, nil
}
