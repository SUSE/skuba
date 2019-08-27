import pytest
import requests


def test_nginx_deployment(deployment, kubectl):
    workers = kubectl.skuba.num_of_nodes("worker")
    kubectl.run_kubectl("create deployment nginx --image=nginx:stable-alpine")
    kubectl.run_kubectl("scale deployment nginx --replicas={replicas}".format(replicas=workers))
    kubectl.run_kubectl("expose deployment nginx --port=80 --type=NodePort")
    kubectl.run_kubectl("wait --for=condition=available deploy/nginx --timeout={timeout}m".format(timeout=3))
    readyReplicas = kubectl.run_kubectl("get deployment/nginx -o jsonpath='{ .status.readyReplicas }'")

    assert int(readyReplicas) == workers

    nodePort = kubectl.run_kubectl("get service/nginx -o jsonpath='{ .spec.ports[0].nodePort }'")

    wrk_idx = 0
    ip_addresses = kubectl.skuba.platform.get_nodes_ipaddrs("worker")

    url = "{protocol}://{ip}:{port}{path}".format(protocol="http",ip=str(ip_addresses[wrk_idx]),port=str(nodePort),path="/")
    r = requests.get(url)

    assert "Welcome to nginx" in r.text

    # Cleanup
    kubectl.run_kubectl("delete --wait --timeout=60s service/nginx")
    kubectl.run_kubectl("delete --wait --timeout=60s deployments/nginx")

    with pytest.raises(Exception):
        kubectl.run_kubectl("get service/nginx")

    with pytest.raises(Exception):
        kubectl.run_kubectl("get deployments/nginx")
