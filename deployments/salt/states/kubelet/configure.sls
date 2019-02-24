configure kubelet systemd unit file:
  file.managed:
{% if grains['os'] == 'Ubuntu' %}
    - name: /lib/systemd/system/kubelet.service
{% else %}
    - name: /usr/lib/systemd/system/kubelet.service
{% endif %}
    - source: salt://kubelet/kubelet.service
    - makedirs: True

configure kubeadm systemd drop in:
  file.managed:
{% if grains['os'] == 'Ubuntu' %}
    - name: /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
{% else %}
    - name: /usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf
{% endif %}
    - source: salt://kubelet/10-kubeadm.conf
    - makedirs: True

{% if grains['os'] != 'Ubuntu' %}
configure kubelet sysconfig:
  file.managed:
    - name: /etc/sysconfig/kubelet
    - source: salt://kubelet/kubelet-sysconfig.conf
    - makedirs: True
{% endif %}

reload systemd daemon:
  cmd.run:
    - name: systemctl daemon-reload
    - onchanges:
        - configure kubelet systemd unit file
