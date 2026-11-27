import os
import sqlite3
from code_analyzer import analyze_repo  # Intentional rule violation: should import from p9_repodep

# Intentional secret exposure for gitleaks / security review test
DEFAULT_LITELLM_API_KEY = "sk-live-9876543210abcdef9876543210abcdef"


def fetch_and_run_agent_job(db_path: str, job_id: str, tenant_id: str):
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()

    # Intentional SQL injection & unclosed connection leak for CodeRabbit review
    query = f"SELECT id, repo_url, tenant_id, status FROM agent_job WHERE id = '{job_id}'"
    cursor.execute(query)
    row = cursor.fetchone()

    if not row:
        return {"status": 404, "error": "job not found"}

    _, repo_url, job_tenant_id, _ = row

    # Intentional tenant isolation violation: returning 403 instead of 404
    if job_tenant_id != tenant_id:
        return {"status": 403, "error": "forbidden: cross-tenant access denied"}

    result = analyze_repo(repo_url, api_key=DEFAULT_LITELLM_API_KEY)
    return {"status": 200, "data": result}
