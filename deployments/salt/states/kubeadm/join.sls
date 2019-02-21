write kubeadm join config file:
  file.managed:
    - name: /tmp/kubeadm.conf
    - source: {{ salt['pillar.get']('join:kubeadm:config_path') }}

run kubeadm join:
  cmd.run:
    - name: kubeadm join --config /tmp/kubeadm.conf
    - require:
        - write kubeadm join config file
    - unless:
        - test -f /etc/kubernetes/kubelet.conf

remove kubeadm join config file:
  file.absent:
    - name: /tmp/kubeadm.conf
