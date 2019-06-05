#!/usr/bin/env python
# -*- encoding: utf-8 -*-

from setuptools import setup
import os

def version():
    """Return the version. Pass when VERSION file not found"""
    pwd = os.path.dirname(os.path.abspath(__file__))
    try:
        with open(os.path.join(pwd, '..', 'VERSION')) as version_file:
            return version_file.read().strip()
    except FileNotFoundError:
        return ''

setup(
    name = "skuba-update",
    version = version(),
    author = "SUSE Containers Team",
    author_email = "containers@suse.com",
    description = "Utility to automatically refresh and update a SUSE CaaSP cluster nodes",
    long_description = "Wraps zypper to refresh repositories and apply non interactive patches, including patches requiring reboot.",
    license = "Apache License 2.0",
    keywords = "CaaSP",
    url = "https://github.com/SUSE/skuba-update",
    packages=['skuba_update'],
    classifiers=[
        'Intended Audience :: Developers',
        'License :: OSI Approved :: Apache License 2.0',
        'Operating System :: POSIX :: Linux',
    ], data_files={
        'skuba_update/skuba-update.timer',
        'skuba_update/skuba-update.service'
    }, entry_points = {
        'console_scripts': [
            'skuba-update = skuba_update.skuba_update:main'
        ]
    }
)
