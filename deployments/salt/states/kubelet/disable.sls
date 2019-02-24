stop and disable the kubelet service:
  service.dead:
    - name: kubelet
    - enable: False
