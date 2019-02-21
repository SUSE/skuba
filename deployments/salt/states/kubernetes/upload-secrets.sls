upload control plane secrets:
  file.recurse:
    - name: /etc/kubernetes/pki
    - source: {{ salt['pillar.get']('join:kubernetes:secrets_path') }}
    - makedirs: True

upload control plane admin kubeconfig file:
  file.managed:
    - name: /etc/kubernetes/admin.conf
    - source: {{ salt['pillar.get']('join:kubernetes:admin_conf_path') }}
