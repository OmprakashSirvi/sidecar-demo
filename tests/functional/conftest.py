import pytest


# returns sidecar host
@pytest.fixture(scope="session")
def sidecar_host():
    return "sidecar"


# Sidecar API URL (for ticket operations)
@pytest.fixture(scope="session")
def sidecar_api_url(sidecar_host) -> str:
    return f"http://{sidecar_host}:8070"


# Backend URL accessed via sidecar proxy
@pytest.fixture(scope="session")
def backend_via_sidecar_url(sidecar_host) -> str:
    return f"http://{sidecar_host}:8080"
