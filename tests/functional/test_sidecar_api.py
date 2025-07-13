import pytest
import requests
from http import HTTPStatus


# @pytest.param
def test_sidecar_apis(sidecar_api_url) -> None:
    res = requests.get(f"{sidecar_api_url}/info")
    assert res.status_code is HTTPStatus.OK
