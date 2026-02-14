# Track Spec: Example Suite using LangChainGo and Comprehensive Documentation

## Overview
Demonstrate the practical utility of RLM-Go by building a bridge to the `langchaingo` ecosystem. Create real-world examples showing how RLM-DSL scripts can be executed by a Go host that uses LangChainGo to interact with real LLM providers.

## Requirements

1. **LangChainGo Integration (`examples/bridge`)**:
    - Implement the `runtime.Host` interface using `github.com/tmc/langchaingo`.
    - Support at least one provider (e.g., Ollama or OpenAI) for the `SUBCALL` implementation.
    - Provide a configurable "RLM Host" that manages the session loop and the LLM interaction logic.

2. **Example RLM-DSL Scripts (`examples/scripts`)**:
    - **`summarize.rlm`**: A script that recursively breaks down a long document and produces a summary.
    - **`extract.rlm`**: A script that searches for specific JSON patterns in a large prompt and parses them.
    - **`router.rlm`**: A script that classifies a prompt and subcalls a specialized task.

3. **Complete Documentation**:
    - **`README.md`**: Update with:
        - Installation and CLI usage.
        - The LangChainGo bridge example.
        - Explanation of the Recursive Language Model paradigm.
    - **`docs/`**: Organize existing specs and protocols into a developer guide.

## Success Criteria
- A developer can run a Go example that uses LangChainGo to execute a recursive RLM-DSL script.
- All examples are well-commented and easy to understand.
- Documentation is accurate, comprehensive, and follows the project style guide.
