# v0.1 Test Fixture Corpus

This document tracks the v0.1 fixtures used as a baseline for regression testing.

## RLM Files (.rlm)
- `test.rlm`: Minimal smoke test.
- `benchmark_patterns.rlm`: Comprehensive test of various features (literals, escapes, safety, recursion, regex).
- `extract.rlm`: Test of extraction patterns.
- `summarize.rlm`: Test of summarization patterns.

## Data Files (.jsonl, .json)
- `suiteA.jsonl`
- `suiteB.jsonl`
- `suiteC.jsonl`
- `suiteC.sample.jsonl`
- `suiteD.jsonl`

## Purpose
These files represent the stable v0.1 surface area. Any changes to the parser or runtime must ensure these files continue to parse and execute correctly in COMPAT mode.
