write cni config file:
  file.managed:
    - name: /tmp/cni.conf
    - source: {{ salt['pillar.get']('bootstrap:cni:config_path') }}

deploy cni:
  cmd.run:
    - name: kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.conf
    - require:
        - write cni config file

remove cni config file:
  file.absent:
    - name: /tmp/cni.conf
    - require:
        - write cni config file
