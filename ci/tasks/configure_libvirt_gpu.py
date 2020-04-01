#!/usr/bin/env python3

import argparse
import logging
import subprocess
import tempfile
import time
import xml.etree.ElementTree as ET

from contextlib import contextmanager

logger = logging.getLogger('Libvirt-Attach')


class ConfigureLibvirtDevice:

    def shutdown_domain(self, domain):
        if 'shut off' not in self._get_domain_status(domain):
            logger.info(f'Shutting down {domain}')
            self._run_cmd(f'virsh shutdown {domain}')

            if not self._wait_for_status(domain, 'shut off'):
                logger.warning(f'{domain} still not off, forcing off')
                self._run_cmd(f'virsh destroy {domain}')

    def detach_device(self, device_id):
        logger.info(f'Detaching {device_id} from host')
        self._run_cmd(f'virsh nodedev-detach {device_id}')

    def attach_device(self, domain, device_id):
        logger.info(f'Attaching {device_id} to {domain}')
        device_addresses = configure._get_device_addresses(device_id)

        if not device_addresses:
            raise Exception(f'Was not able to retrieve device addresses for {device_id}')

        with configure._write_config_file(device_addresses) as config_file:
            self._run_cmd(f'virsh attach-device {domain} --file {config_file.name} --config')

    def start_domain(self, domain):
        logger.info(f'Starting {domain}')
        self._run_cmd(f'virsh start {domain}')

    def _get_device_addresses(self, device_id):
        output = self._run_cmd(f'virsh nodedev-dumpxml {device_id}')
        root = ET.fromstring(output)

        return root.findall("./capability/iommuGroup/address")

    def _get_domain_status(self, domain):
        return self._run_cmd(f'virsh domstate {domain}')

    def _run_cmd(self, cmd):
        logger.debug(cmd)

        proc = subprocess.run(cmd,
                              encoding='utf8',
                              shell=True,
                              stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT)
        if proc.returncode != 0:
            raise Exception(f'Received exit code {proc.returncode} while running command {cmd}\n{proc.stdout}')

        return proc.stdout

    def _wait_for_status(self, domain, status, timeout=60):
        current_status = self._get_domain_status(domain)

        while status not in current_status and timeout > 0:
            timeout -= 1
            time.sleep(1)
            current_status = self._get_domain_status(domain)

        return status in current_status

    @contextmanager
    def _write_config_file(self, device_addresses):
        xml_doc = ET.Element('hostdev', attrib={'mode': 'subsystem', 'type': 'pci', 'managed': 'yes'})
        source = ET.SubElement(xml_doc, 'source')
        source.extend(device_addresses)
        body = str(ET.tostring(xml_doc), 'utf-8')

        with tempfile.NamedTemporaryFile('w', encoding='utf-8') as f:
            logging.debug(f'Config file path {f.name}')
            logging.debug(f'Config file content {body}')
            f.write(body)
            f.flush()
            yield f


def define_parser(parser):
    parser.add_argument('domain',
                        help='Name of the domain to attach the device to')

    parser.add_argument('device_id',
                        help='ID of the device to attach e.g. pci_0000_03_00_0')

    parser.add_argument('--debug', action='store_true',
                        help='Enable debugging output')


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Attach a host device to a libvirt VM')
    define_parser(parser)
    args = parser.parse_args()

    logging.basicConfig(format='%(asctime)s %(name)s: %(levelname)s: %(message)s', level='DEBUG' if args.debug else 'INFO')

    configure = ConfigureLibvirtDevice()

    configure.shutdown_domain(args.domain)
    configure.detach_device(args.device_id)
    configure.attach_device(args.domain, args.device_id)
    configure.start_domain(args.domain)
