import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path
import google.generativeai as genai

# Setup Gemini
api_key = os.getenv("GOOGLE_API_KEY") or os.getenv("GEMINI_API_KEY")
if not api_key:
    print("Error: GOOGLE_API_KEY not set")
    sys.exit(1)

genai.configure(api_key=api_key)
model = genai.GenerativeModel('gemini-2.0-flash')

def run_python_task(case_id, task, prompt_content, expected_val):
    # Prompt the LLM to generate Python code
    llm_prompt = f"""
Task: {task}
Context/Data:
\"\"\"
{prompt_content}
\"\"\"

Write a short Python script that solves this task and prints the result as JSON to stdout.
Do not include any explanation, just the code.
If the task asks for a specific value from the data, extract it and print it.
"""
    
    response = model.generate_content(llm_prompt)
    code = response.text.strip()
    if "```python" in code:
        code = code.split("```python")[1].split("```")[0].strip()
    elif "```" in code:
        code = code.split("```")[1].split("```")[0].strip()

    # Execute code in a temp file using uv
    with tempfile.NamedTemporaryFile(suffix=".py", mode="w", delete=False) as f:
        f.write(code)
        temp_path = f.name

    try:
        # Run with uv run
        result = subprocess.run(
            ["uv", "run", temp_path],
            capture_output=True,
            text=True,
            timeout=5
        )
        
        output_str = result.stdout.strip()
        try:
            actual_val = json.loads(output_str)
        except:
            actual_val = output_str

        # Simple semantic comparison
        passed = False
        if expected_val is None:
            passed = result.returncode == 0
        elif isinstance(expected_val, (int, float)) and str(actual_val) == str(expected_val):
            passed = True
        elif str(actual_val).strip() == str(expected_val).strip():
            passed = True
        elif isinstance(expected_val, dict) and isinstance(actual_val, dict):
            passed = all(actual_val.get(k) == v for k, v in expected_val.items())
        
        return passed, actual_val, code
    except Exception as e:
        return False, str(e), code
    finally:
        if os.path.exists(temp_path):
            os.remove(temp_path)

def run_suite(suite_file):
    print(f"\nRunning Python Suite: {suite_file.name}")
    results = []
    with open(suite_file, "r") as f:
        for line in f:
            case = json.loads(line)
            case_id = case["id"]
            task = case["task"]
            
            # Load prompt
            prompt_path = Path("bench") / case["prompt_ref"]
            with open(prompt_path, "r") as pf:
                prompt_content = pf.read()
            
            # Load expected
            expected_val = None
            if "expected_ref" in case:
                exp_path = Path("bench") / case["expected_ref"]
                with open(exp_path, "r") as ef:
                    expected_val = json.load(ef)
            
            passed, actual, code = run_python_task(case_id, task, prompt_content, expected_val)
            results.append(passed)
            status = "PASSED" if passed else "FAILED"
            print(f"  [{status}] Case {case_id}: {task}")
            if not passed:
                print(f"    Expected: {expected_val}")
                print(f"    Got: {actual}")
    
    return results

def main():
    suites = sorted(list(Path("bench/cases").glob("suite*.jsonl")))
    summary = {}
    
    for suite in suites:
        # Filter suites to match the ones user highlighted
        if suite.name not in ["suiteA.jsonl", "suiteB.jsonl", "suiteC.jsonl", "suiteE.jsonl", "suiteF.jsonl"]:
            continue
            
        res = run_suite(suite)
        summary[suite.name] = (len(res), sum(res))

    print("\n\nPYTHON VS RLM-DSL COMPARISON")
    print("-------------------------------------------------------")
    print(f"{'Suite':<20} {'Total':<10} {'Passed':<10} {'Success %':<10}")
    print("-" * 55)
    for name, (total, passed) in summary.items():
        pct = (passed / total) * 100
        print(f"{name:<20} {total:<10} {passed:<10} {pct:<10.1f}%")

if __name__ == "__main__":
    main()
