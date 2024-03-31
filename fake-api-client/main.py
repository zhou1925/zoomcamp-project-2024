import requests
from faker import Faker
import json
import time

# Initialize Faker
fake = Faker()

# Infinite loop to send data 
while True:
    # Generate random data
    insights_data = {
        "top_job_skills": {fake.job(): fake.random_int(min=1, max=10) for _ in range(5)},
        "top_companies": {fake.company(): fake.random_int(min=1, max=10) for _ in range(5)},
        "top_cities": {fake.city(): fake.random_int(min=1, max=10) for _ in range(5)},
        "top_states": {fake.state(): fake.random_int(min=1, max=10) for _ in range(5)}
    }

    # Convert data to JSON
    payload = json.dumps({"insights": insights_data})

    # Send POST request
    # url = 'http://localhost:8080/produce'
    url = 'http://producer:8080/produce'  # Change 'localhost' to 'producer'
    headers = {'Content-Type': 'application/json'}
    response = requests.post(url, headers=headers, data=payload)

    # Check response
    if response.status_code == 200:
        print("Message sent successfully")
    else:
        print("Failed to send message. Status code:", response.status_code)
        print("Response:", response.text)

    # Wait for interval
    # 1 hour
    interval = 1 * 60 * 60
    interval = 1
    time.sleep(interval)
