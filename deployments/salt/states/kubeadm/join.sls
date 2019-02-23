write kubeadm join config file:
  file.managed:
    - name: /tmp/kubeadm.conf
    - source: {{ salt['pillar.get']('join:kubeadm:config_path') }}

enable cri:
  service.running:
    - name: docker
    - enable: True

stop kubelet service:
  service.dead:
    - name: kubelet

run kubeadm join:
  cmd.run:
    - name: kubeadm join --config /tmp/kubeadm.conf
    - require:
        - write kubeadm join config file
        # kubeadm sanity checks will fail if the CRI is not enabled at next reboot
        - enable cri
        # ensure the kubelet service is not running, could cause kubeadm sanity checks to fail
        - stop kubelet service
    - unless:
        - test -f /etc/kubernetes/kubelet.conf

remove kubeadm join config file:
  file.absent:
    - name: /tmp/kubeadm.conf
