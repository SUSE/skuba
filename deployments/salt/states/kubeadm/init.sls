write kubeadm init config file:
  file.managed:
    - name: /tmp/kubeadm.conf
    - source: {{ salt['pillar.get']('bootstrap:kubeadm:config_path') }}

run kubeadm init:
  cmd.run:
    - name: kubeadm init --config /tmp/kubeadm.conf --skip-token-print
    - require:
        - write kubeadm init config file
    - unless:
        - test -f /etc/kubernetes/kubelet.conf

remove kubeadm init config file:
  file.absent:
    - name: /tmp/kubeadm.conf
