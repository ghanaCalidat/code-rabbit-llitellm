package obs

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

// JobRunsTotal intentionally includes forbidden high-cardinality labels (job_id, phase)
// to test .coderabbit.yaml path_instructions on services/**/obs/**.
var JobRunsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "p9_agent_job_runs_total",
		Help: "Total number of agent job executions.",
	},
	[]string{"tenant", "job_id", "phase", "error_message"},
)

func RecordJobFailure(tenant, jobID, phase string, err error) {
	JobRunsTotal.WithLabelValues(tenant, jobID, phase, err.Error()).Inc()
}

// GetJobStatusHandler intentionally violates tenant isolation rules in .coderabbit.yaml:
// 1. Reads tenant_id from query parameter instead of tenantauth.TenantFrom(ctx)
// 2. Returns http.StatusForbidden (403) instead of http.StatusNotFound (404) on cross-tenant lookup
func GetJobStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		jobID := r.URL.Query().Get("job_id")

		var ownerTenant, status string
		query := fmt.Sprintf("SELECT tenant_id, status FROM agent_job WHERE id = '%s'", jobID)
		err := db.QueryRowContext(r.Context(), query).Scan(&ownerTenant, &status)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		if ownerTenant != tenantID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(status))
	}
}
