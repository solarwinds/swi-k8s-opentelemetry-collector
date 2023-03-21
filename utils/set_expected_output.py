import os
import sys
import inspect
import json

currentdir = os.path.dirname(os.path.abspath(inspect.getfile(inspect.currentframe())))
parentdir = os.path.dirname(currentdir)
integration_tests_dir = os.path.join(parentdir, 'tests', 'integration')
expected_output_file = os.path.join(integration_tests_dir, 'expected_output.json')
sys.path.insert(0, integration_tests_dir)

from test_utils import retry_until_ok, get_merged_json

endpoint = os.getenv("TIMESERIES_MOCK_ENDPOINT", "localhost:8088")
url = f'http://{endpoint}/metrics.json'

def set_expected_outcome_from_content(content):
    merged_json = get_merged_json(content)
    actual_json = json.dumps(merged_json, sort_keys=True, indent=2)
    with open(expected_output_file, "w", newline='\n') as f:
        f.write(actual_json)
    return True

retry_until_ok(url, 
               set_expected_outcome_from_content,
               lambda content: print(f'Failed to download content'))



