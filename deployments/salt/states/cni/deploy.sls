write cni config dir:
  file.recurse:
    - name: /tmp/cni.conf.d
    - source: {{ salt['pillar.get']('bootstrap:cni:config_dir') }}

deploy cni:
  cmd.run:
    - name: kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /tmp/cni.conf.d
    - require:
        - write cni config dir

remove cni config dir:
  file.absent:
    - name: /tmp/cni.conf.d
