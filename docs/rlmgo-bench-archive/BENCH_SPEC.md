
# EnvLLM Benchmark Suite Specification
Version: 0.1

This suite proves:
1) DSL reliability (parse + validate success)
2) VM correctness and safety
3) RLM effectiveness on long prompts
4) Deterministic, schema-validated observations

Suites:
A — DSL Reliability
B — VM Correctness & Budgets
C — Long-Context RLM Tasks
D — Tool-Orchestration / AST scoring (optional)

All cases use JSONL envelopes and external prompt/expected files.
