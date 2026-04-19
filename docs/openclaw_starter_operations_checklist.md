# OpenClaw Starter Operations Checklist

This checklist is the operator handoff for the 30-day OpenClaw strategy launch.

## Day 0 Setup

- [ ] Confirm strategy file exists: `openclaw_strategy_agents.yaml`
- [ ] Confirm source strategy PDF path is accessible:
  - `C:\Users\Keith Ransom\Downloads\openclaw-business-strategy.pdf`
- [ ] Confirm model provider keys are available in environment.
- [ ] Run baseline kickoff command:
  - `./bin/cagent exec ./openclaw_strategy_agents.yaml "Read C:\Users\Keith Ransom\Downloads\openclaw-business-strategy.pdf with analyze_document, then produce a 30-day execution plan with assigned sub-agent owners and weekly milestones."`

## Weekly Commands

Use the built-in command shortcuts from the root agent:

- [ ] Week 1: `./bin/cagent run ./openclaw_strategy_agents.yaml -c week1`
- [ ] Week 2: `./bin/cagent run ./openclaw_strategy_agents.yaml -c week2`
- [ ] Week 3: `./bin/cagent run ./openclaw_strategy_agents.yaml -c week3`
- [ ] Week 4: `./bin/cagent run ./openclaw_strategy_agents.yaml -c week4`

## KPI Tracking

Track these KPIs weekly:

- [ ] Skills published
- [ ] MSP pilots / signed LOIs
- [ ] Enterprise discovery engagements
- [ ] Monthly revenue run-rate and forecast delta

## Exit Criteria for First 30 Days

- [ ] Two skills are published and stable.
- [ ] One MSP pilot is active or signed.
- [ ] At least one enterprise discovery process is completed.
- [ ] A 60-day follow-on plan is documented with owners and dates.
