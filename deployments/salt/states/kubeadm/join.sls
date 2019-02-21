write kubeadm join config file:
  file.managed:
    - name: /tmp/kubeadm.conf
    - source: {{ salt['pillar.get']('kubeadm:config_path') }}

run kubeadm join:
  cmd.run:
    - name: kubeadm join --config /tmp/kubeadm.conf
    - require:
        - write kubeadm join config file

remove kubeadm join config file:
  file.absent:
    - name: /tmp/kubeadm.conf
