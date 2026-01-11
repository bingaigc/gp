# Simple keep-alive script for Databricks container service
# The actual service runs in the Docker container

import time
import requests
import sys

def check_service_health():
    """Check if the containerized service is healthy"""
    try:
        response = requests.get("http://localhost:8080/health", timeout=5)
        return response.status_code == 200
    except Exception as e:
        print(f"Health check failed: {e}")
        return False

def main():
    print("Starting Alpha Detector Service monitor...")
    print("Service is running in Docker container")
    print("API available at: http://localhost:8080")
    
    # Keep the Databricks job alive while monitoring the service
    failure_count = 0
    max_failures = 5
    
    while True:
        if check_service_health():
            print("✓ Service is healthy")
            failure_count = 0
        else:
            failure_count += 1
            print(f"✗ Service health check failed ({failure_count}/{max_failures})")
            
            if failure_count >= max_failures:
                print("Service has failed too many times. Exiting.")
                sys.exit(1)
        
        # Check every 30 seconds
        time.sleep(30)

if __name__ == "__main__":
    main()
