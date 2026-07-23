---
type: discovery-brief
id: 2026-alert-triage-queues
title: Priority queues for alert triage
status: validated
team: fraud-detection
created: 2026-07-08
tags: [kind/discovery, status/validated, team/fraud-detection]
---

# Discovery: Priority queues for alert triage

## Problem signal
Median time-to-disposition for high-risk Alerts was 6.4h in Q2 (SLA: 2h).
Analysts triage a single FIFO queue; 71% of the queue is low-risk noise.

## Hypothesis
We believe risk-ranked triage queues will bring high-risk median
time-to-disposition under 2h without increasing false-negative rate.

## Success criteria
- High-risk median time-to-disposition < 2h within 30 days.
- False-negative rate unchanged (±0.1pp).
- Analyst context switches per shift down 40%.

## Affected components (initial guess)
- `transaction-screening`

## Risks and open questions
- Ranking model drift; weekly calibration owned by the team.

## Decision
validated — 2026-07-09, quantified signal, numeric criteria.
