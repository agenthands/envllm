
EnvLLM Benchmark Starter

Contains:
- BENCH_SPEC.md — benchmark definition
- schemas/ — JSON schemas
- cases/ — JSONL test cases
- prompts/ — large prompt files
- expected/ — expected outputs
- runner/ — Go harness skeleton

Usage:
1) implement Model adapter
2) implement runtime execution
3) run cases and produce report JSON
