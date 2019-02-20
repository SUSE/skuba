kubeadm join --config {{ salt['pillar.get']('kubeadm:config_path') }}:
  cmd.run
