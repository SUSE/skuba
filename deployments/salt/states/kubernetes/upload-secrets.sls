upload control plane secrets:
  file.recurse:
    - name: /etc/kubernetes/pki
    - source: {{ salt['pillar.get']('join:kubernetes:secrets_path') }}
