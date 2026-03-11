---
name: batch-ci-analysis
argument-hint: <url1> <url2> [url3] ...
description: Analyze multiple Prow CI jobs in parallel and produce a collated report. Use proactively when the user provides multiple prow.ci.openshift.org URLs.
allowed-tools: WebFetch, Bash, Read, Write, Glob, Grep
---

# Batch CI Analysis

Analyze multiple Prow CI job URLs in parallel using sub-agents and produce a single collated human-readable report.

## Input

URLs to analyze:
$ARGUMENTS

## Workflow

This analysis uses a tiered approach:
- **Tier 1**: Quick structured overview of ALL jobs (parallel)
- **Tier 2**: Deep error analysis of FAILED jobs only (parallel)
- **Collation**: Unified report with summary table, cross-job patterns, and failure details

### Phase 1: Parse and Validate URLs

1. Extract all URLs from the input. URLs match the Prow CI pattern:
   - `https://prow.ci.openshift.org/view/gs/test-platform-results/logs/<job-name>/<job-id>`
   - `https://prow.ci.openshift.org/view/gs/test-platform-results/pr-logs/pull/openshift_microshift/<pr-number>/<job-name>/<job-id>`
   - URLs may be separated by spaces, newlines, commas, or embedded in markdown/text
2. Validate each URL contains `prow.ci.openshift.org`
3. Report any invalid URLs to the user and continue with valid ones
4. If zero valid URLs found, stop and ask the user for valid URLs
5. Assign each URL a short identifier (Job 1, Job 2, etc.)

### Phase 2: Tier 1 — Structured Overview (All Jobs, Parallel)

For EACH valid URL, spawn an `openshift-ci-analysis` sub-agent using the Task tool.

**CRITICAL**: Make ALL Task tool calls in a SINGLE message so they run in parallel.

Use these parameters for each Task call:
- `subagent_type`: `openshift-ci-analysis`
- `description`: `Overview CI job N` (where N is the job number)
- `max_turns`: 15
- `prompt`: Use the template below, substituting the URL:

> Analyze this Prow CI job and return ONLY a concise structured summary.
> Do NOT perform deep error analysis. Do NOT scan individual log files for errors.
> Focus on extracting metadata and high-level pass/fail status.
>
> URL: {URL}
>
> Steps:
> 1. Fetch `finished.json` to get job result and timing
> 2. Fetch `started.json` for start time
> 3. Extract the job name, job ID, version, and architecture from the URL
> 4. List all test scenarios from the scenario-info directory
> 5. For each scenario, check only the junit.xml for pass/fail counts
>
> Return your findings in EXACTLY this format (one field per line):
>
> ```
> JOB_STATUS: SUCCESS|FAILURE|ABORTED
> JOB_NAME: <full job name>
> JOB_ID: <numeric job id>
> VERSION: <MicroShift version or "unknown">
> ARCH: <x86_64|aarch64>
> IMAGE_TYPE: <bootc|rpm-ostree|unknown>
> DURATION: <e.g. "1h 23m 45s">
> STARTED: <YYYY-MM-DD HH:MM:SS UTC>
> FINISHED: <YYYY-MM-DD HH:MM:SS UTC>
> TOTAL_SCENARIOS: <count>
> PASSED_SCENARIOS: <count>
> FAILED_SCENARIOS: <count>
> FAILED_SCENARIO_NAMES: <comma-separated list, or "none">
> ERROR_SUMMARY: <one-line cause of failure, or "N/A">
> PROW_URL: <the original prow URL>
> ```

### Phase 3: Tier 2 — Deep Error Analysis (Failed Jobs Only, Parallel)

After ALL Tier 1 sub-agents complete:

1. Parse each Tier 1 result to identify jobs with `JOB_STATUS: FAILURE`
2. If there are NO failures, skip to Phase 4
3. For each failed job, spawn an `openshift-ci-analysis` sub-agent:

**CRITICAL**: Make ALL Tier 2 Task tool calls in a SINGLE message so they run in parallel.

Use these parameters:
- `subagent_type`: `openshift-ci-analysis`
- `description`: `Deep analysis failed job N`
- `max_turns`: 30
- `prompt`: Use the template below, substituting the URL:

> Perform a thorough error analysis on this failed Prow CI job.
>
> URL: {URL}
>
> Follow the full analysis workflow:
> 1. Create a temp directory with `mktemp -d /tmp/openshift-ci-analysis-XXXX`
> 2. Fetch the top-level build-log.txt
> 3. Scan for errors and record each error with filepath and line number
> 4. For each error, read 50 lines before and after for context
> 5. Classify each error: root cause vs transient vs red herring
> 6. Determine the stack layer: AWS infra, build phase, deploy phase, test, teardown
> 7. If the error is a legitimate test failure, determine what stage failed (setup, testing, teardown)
> 8. If the source appears to be MicroShift-related, check the SOS report's journal and pod logs
>
> Return your analysis using this format for EACH distinct error found:
>
> ```
> Error Severity: {1-5}
> Stack Layer: {AWS Infra | Build Phase | Deploy Phase | Test | Teardown}
> Step Name: {the specific CI step where the error occurred}
> Error: {the exact error with relevant log context}
> Suggested Remediation: {actionable fix or "Infrastructure - no code change needed"}
> ```
>
> After all errors, provide a one-paragraph ROOT CAUSE SUMMARY explaining what
> ultimately caused this job to fail.

### Phase 4: Collate and Report

After ALL sub-agents from both tiers complete, produce the final report.

**Report Template**:

```markdown
# Batch CI Analysis Report

**Date**: {current date}
**Jobs Analyzed**: {count}
**Passed**: {count} | **Failed**: {count} | **Aborted**: {count}

## Job Overview

| # | Status | Job Name | Version | Arch | Duration | Failed Scenarios |
|---|--------|----------|---------|------|----------|------------------|
| 1 | ✅/❌/⚠️ | <name>  | <ver>   | <arch> | <dur>  | <names or "—">   |

## Cross-Job Patterns

{Analyze the results across ALL jobs and answer:}
- Are multiple jobs failing with the same root cause?
- Are failures infrastructure-related (AWS, quota, hypervisor) or test-related?
- Is a specific MicroShift version or scenario consistently failing?
- Are there common flaky tests appearing across jobs?

## Failed Job Details

### ❌ Job {N}: {job-name}

- **Prow**: [View on Prow]({url})
- **Status**: FAILURE
- **Version**: {version}
- **Failed Scenarios**: {list}
- **Root Cause**: {one-line summary}

#### Error Analysis

{Full error report from Tier 2 sub-agent} 

{Repeat for each failed job}

## Passing Jobs

| # | Job Name | Version | Arch | Duration | Scenarios |
|---|----------|---------|------|----------|-----------|
{Table of passing jobs}
```

## Error Handling

- If a sub-agent fails or times out, include a note in the report and continue with remaining jobs
- If a URL is inaccessible (404, network error), report it in the summary and skip that job
- If Tier 1 cannot determine pass/fail status for a job, include it in Tier 2 analysis as a precaution
- If all sub-agents fail, report the failures and suggest the user try individual analysis with `/analyze-ci-test-job`

## Notes

- This command spawns `openshift-ci-analysis` sub-agents via the Task tool
- Tier 1 sub-agents are limited to 15 turns for fast metadata extraction
- Tier 2 sub-agents get 30 turns for thorough error investigation
- All sub-agents within a tier run in parallel (multiple Task calls in one message)
- For best results, provide 2-10 URLs; more than 10 may exceed context limits
