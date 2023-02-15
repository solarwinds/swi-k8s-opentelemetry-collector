import json
import time
import requests

def get_all_bodies(log_bulk):
    result = [records["body"]["stringValue"]
              for resource in log_bulk["resourceLogs"]
              for scope in resource["scopeLogs"]
              for records in scope["logRecords"]
              ]
    return result

def get_all_bodies_for_all_sent_content(content):
    lines = content.splitlines()
    log_bulks = [json.loads(line) for line in lines]
    return [get_all_bodies(log_bulk) for log_bulk in log_bulks]

def retry_until_ok(url, func, print_failure):
    timeout = 120  # set the timeout in seconds
    start_time = time.time()
    last_exception = None
    while time.time() - start_time < timeout:
        is_ok = False
        try: 
            try: 
                response = requests.get(url)
                response.raise_for_status()
            except requests.exceptions.RequestException as e:
                print(f"An error occurred while making the request: {e}")

            if response is not None and response.status_code == 200:
                print("Successfully downloaded!")
                is_ok = func(response.content)
            else:
                if response is not None:
                    print('Failed to download otel messages. Response code:',
                        response.status_code)

                return False
        except Exception as e:
            last_exception = e
            print('An exception occurred: {}'.format(e))

        if is_ok:
            print(f'Succesfully passed assert')
            break
        else:
            print('Retrying...')
            time.sleep(2)

    if time.time() - start_time >= timeout:
        if last_exception is not None:
            print('Last exception: {}'.format(last_exception))
        
        if response is not None:
            print_failure(response.content)

        raise ValueError("Timed out waiting")
