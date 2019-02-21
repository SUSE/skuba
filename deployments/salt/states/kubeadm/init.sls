write kubeadm init config file:
  file.managed:
    - name: /tmp/kubeadm.conf
    - source: {{ salt['pillar.get']('kubeadm:config_path') }}

run kubeadm init:
  cmd.run:
    - name: kubeadm init --config /tmp/kubeadm.conf
    - require:
        - write kubeadm init config file

remove kubeadm init config file:
  file.absent:
    - name: /tmp/kubeadm.conf
