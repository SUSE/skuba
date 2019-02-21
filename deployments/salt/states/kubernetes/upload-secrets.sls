upload control plane secrets:
  file.recurse:
    - name: /etc/kubernetes
    - source: {{ salt['pillar.get']('join:kubernetes:secrets_path') }}
