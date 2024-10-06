"""
Pytest suite to perform end-to-end-testing on the reverse proxy.
"""
from typing import Dict, List
from threading import Thread
from http.server import SimpleHTTPRequestHandler, HTTPServer
import subprocess
import time
import pytest
import requests

MOCK_SERVER_RESPONSE = "mock server resonse\n".encode()

class ProxyRequestHandler(SimpleHTTPRequestHandler):
    """
    Request handler used to log requests made by a reverse proxy instance in E2E tests.
    """

    # Maps string paths to the number of times they've been requested
    path_requests_counter: Dict[str, int] = {}
    
    def do_GET(self):
        if self.path in ProxyRequestHandler.path_requests_counter:
            ProxyRequestHandler.path_requests_counter[self.path] += 1
        else:
            ProxyRequestHandler.path_requests_counter[self.path] = 1

        self.send_response(200)
        self.send_header("Content-type", "text/plain")
        self.end_headers()
        self.wfile.write(MOCK_SERVER_RESPONSE)

class MockServers:
    def __init__(self, ports: List[int]):
        self._ports = ports
        self._servers: List[HTTPServer] = []

    def _mock_server_impl(self, port):
        server = HTTPServer(("localhost", port), ProxyRequestHandler)
        self._servers.append(server)
        server.serve_forever()

    def __enter__(self):
        ProxyRequestHandler.path_requests_counter = {}

        for port in self._ports:
            print(f"Starting server on port {port}")
            # Use daemon=True to ensure the servers die after the main process exits
            thread = Thread(target=lambda: self._mock_server_impl(port), daemon=True)
            thread.start()

    def __exit__(self, exc_type, exc_val, exc_tb):
        for server in self._servers:
            server.shutdown()

        ProxyRequestHandler.path_requests_counter = {}

class ReverseProxy:
    def __init__(self, config_path: str):
        self._config_path = config_path

    def __enter__(self):
        self._proxy_process = subprocess.Popen(["./build/reverse-proxy", f"{self._config_path}"])
        time.sleep(5)
        self._proxy_process.poll()
        assert self._proxy_process.returncode is None

    def __exit__(self, exc_type, exc_val, exc_tb):
        self._proxy_process.kill()

def get_metrics(host: str) -> Dict[str, str]:
    """
    Helper function to get and parse Prometheus metrics from the reverse proxy.

    Args:
        host: The hostname and port combo to request metrics from.

    Returns:
        A dict mapping a metrics to its value as it was when calling this function.
    """
    metrics_res = requests.get(f"{host}/metrics").content

    metrics = {}

    for line in metrics_res.decode().split("\n"):
        # Ignore comment lines
        if line.startswith("#"):
            continue

        # Ignore blank lines
        if line.strip() == "":
            continue

        # Name and value are separated by a space
        metric_name, metric_value = line.split(" ")
        metrics[metric_name] = metric_value

    return metrics

def test_mock_servers():
    """
    Ensures that mock servers will work in the current environment.
    """
    with MockServers([5000, 5001]):
        assert requests.get("http://localhost:5000").content == MOCK_SERVER_RESPONSE
        assert requests.get("http://localhost:5001").content == MOCK_SERVER_RESPONSE

def test_routing():
    """
    Tests basic routing of requests and responses through the reverse proxy.
    """
    with MockServers([5000, 5001]), ReverseProxy("e2e/configs/test_routing.yaml"):
        # First, ensure the endpoints haven't been hit
        assert "/index-proxied" not in ProxyRequestHandler.path_requests_counter
        assert "/home-proxied" not in ProxyRequestHandler.path_requests_counter

        metrics = get_metrics("http://localhost:8000")
        assert metrics["index_request_count"] == "0"
        assert metrics["index_last_response_time"] == "0"
        assert metrics["index_successful_request_count"] == "0"
        assert metrics["index_failed_request_count"] == "0"
        assert metrics["home_request_count"] == "0"
        assert metrics["home_last_response_time"] == "0"
        assert metrics["home_successful_request_count"] == "0"
        assert metrics["home_failed_request_count"] == "0"

        # Hit /index (which maps to /index-proxied), and ensure that it was the only endpoint hit
        assert requests.get("http://localhost:8000/index").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied"] == 1
        assert "/home-proxied" not in ProxyRequestHandler.path_requests_counter

        metrics = get_metrics("http://localhost:8000")
        assert metrics["index_request_count"] == "1"
        assert metrics["index_last_response_time"] != "0"
        assert metrics["index_successful_request_count"] == "1"
        assert metrics["index_failed_request_count"] == "0"
        assert metrics["home_request_count"] == "0"
        assert metrics["home_last_response_time"] == "0"
        assert metrics["home_successful_request_count"] == "0"
        assert metrics["home_failed_request_count"] == "0"
        

        # Now ensure we can proxy to the other server as well
        assert requests.get("http://localhost:8000/home").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied"] == 1
        assert ProxyRequestHandler.path_requests_counter["/home-proxied"] == 1

        metrics = get_metrics("http://localhost:8000")
        assert metrics["index_request_count"] == "1"
        assert metrics["index_last_response_time"] != "0"
        assert metrics["index_successful_request_count"] == "1"
        assert metrics["index_failed_request_count"] == "0"
        assert metrics["home_request_count"] == "1"
        assert metrics["home_last_response_time"] != "0"
        assert metrics["home_successful_request_count"] == "1"
        assert metrics["home_failed_request_count"] == "0"

def test_load_balancing():
    """
    Tests load balancing between multiple servers.
    """
    with MockServers([5000, 5001, 5002]), ReverseProxy("e2e/configs/test_load_balancing.yaml"):
        assert "/index-proxied-5000" not in ProxyRequestHandler.path_requests_counter
        assert "/index-proxied-5001" not in ProxyRequestHandler.path_requests_counter
        assert "/index-proxied-5002" not in ProxyRequestHandler.path_requests_counter

        # Ensure at first we go to the 5000 endpoint
        assert requests.get("http://localhost:8000").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5000"] == 1
        assert "/index-proxied-5001" not in ProxyRequestHandler.path_requests_counter
        assert "/index-proxied-5002" not in ProxyRequestHandler.path_requests_counter
        
        # Ensure next request goes to 5001
        assert requests.get("http://localhost:8000").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5000"] == 1
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5001"] == 1
        assert "/index-proxied-5002" not in ProxyRequestHandler.path_requests_counter

        # Ensure next request goes to 5002
        assert requests.get("http://localhost:8000").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5000"] == 1
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5001"] == 1
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5002"] == 1
        
        # Ensure next request wraps back around to 5000
        assert requests.get("http://localhost:8000").content == MOCK_SERVER_RESPONSE
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5000"] == 2
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5001"] == 1
        assert ProxyRequestHandler.path_requests_counter["/index-proxied-5002"] == 1

def test_bad_gateway():
    """
    Tests that the reverse proxy gracefully handles being unable to hit downstream endpoints.
    """

    # We won't spin up any servers here so we know we won't get a response
    with ReverseProxy("e2e/configs/test_bad_gateway.yaml"):
        assert get_metrics("http://localhost:8000")["ROOT_failed_request_count"] == "0"
        assert requests.get("http://localhost:8000").status_code == 502 # Bad gateway
        assert get_metrics("http://localhost:8000")["ROOT_failed_request_count"] == "1"


