#!/usr/bin/env python3

import argparse
import logging
import os
import subprocess
import tarfile
import time

logger = logging.getLogger('Sonobuoy-E2E-Tests')


class SonobuoyE2eTests:
    def __init__(self, artifacts_dir, kubeconfig, sonobuoy_image, sonobuoy_version):
        self.artifacts_dir = artifacts_dir if os.path.isabs(artifacts_dir) else os.path.join(os.getcwd(), artifacts_dir)
        self.default_sleep = 5
        self.kubeconfig = kubeconfig
        self.sonobuoy_image = sonobuoy_image
        self.sonobuoy_version = sonobuoy_version

        if not os.path.isdir(self.artifacts_dir):
            os.mkdir(self.artifacts_dir)

        if not os.path.isfile(kubeconfig):
            raise SonobuoyE2eTestsError(f'No kubeconfig file found at {kubeconfig}')

    def cleanup_cluster(self, sonobuoy_args):
        logger.info(f'Removing sonobuoy tests from the cluster')
        delete = self._sonobuoy('delete --all --wait ' + ' '.join(sonobuoy_args))
        logger.info(delete)

    def collect_results(self, retries, sonobuoy_args):
        attempts = 0

        results = None

        while retries > attempts:
            try:
                attempts += 1
                logger.info(f'Attempting to retrieve the results {attempts}/{retries}')
                results = self._sonobuoy('retrieve results' + ' '.join(sonobuoy_args))
                break
            except SonobuoyE2eTestsError:
                if retries == attempts:
                    raise SonobuoyE2eTestsError('Could not retrieve sonobuoy results')
                time.sleep(self.default_sleep)

        self._extract_results(results.strip())

    def run_tests(self, sonobuoy_args):
        logger.info('Getting the sonobuoy image...')
        self._pull_image()

        logger.info('Starting the tests, they can take up to 2-3 hours')
        self._start_the_tests(sonobuoy_args)

    def _extract_results(self, results_path):
        if tarfile.is_tarfile(results_path):
            results_tar = tarfile.open(results_path)
            results_tar.extractall(self.artifacts_dir)
            os.remove(results_path)
        else:
            raise SonobuoyE2eTestsError(f'Could not extract results from {results_path}')

    def _pull_image(self):
        cmd = f'docker pull {self.sonobuoy_image}:{self.sonobuoy_version}'
        return self._run_cmd(cmd)

    def _run_cmd(self, cmd):
        logger.info(cmd)

        proc = subprocess.run(cmd,
                              encoding='utf8',
                              shell=True,
                              stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT)
        if proc.returncode != 0:
            raise SonobuoyE2eTestsError(f'Received exit code {proc.returncode} while running command {cmd}\n{proc.stdout}')

        return proc.stdout

    def _sonobuoy(self, sonobuoy_args):
        cmd = (f'docker run --rm --network=host '
               f'-v {self.kubeconfig}:/root/.kube/config '
               f'-v {self.artifacts_dir}:/results '
               f'-i {self.sonobuoy_image}:{self.sonobuoy_version} '
               f'./sonobuoy {sonobuoy_args}')
        return self._run_cmd(cmd)

    def _start_the_tests(self, sonobuoy_args):
        self._sonobuoy('run ' + ' '.join(sonobuoy_args) + '--wait')


class SonobuoyE2eTestsError(Exception):
    pass


def run_tests(args, sonobuoy_args):
    sonobuoy_e2e = SonobuoyE2eTests(args.artifacts,
                                    args.kubeconfig,
                                    args.sonobuoy_image,
                                    args.sonobuoy_version)

    sonobuoy_e2e.run_tests(sonobuoy_args)


def collect_results(args, sonobuoy_args):
    sonobuoy_e2e = SonobuoyE2eTests(args.artifacts,
                                    args.kubeconfig,
                                    args.sonobuoy_image,
                                    args.sonobuoy_version)
    sonobuoy_e2e.collect_results(args.collection_retries, sonobuoy_args)


def cleanup(args, sonobuoy_args):
    sonobuoy_e2e = SonobuoyE2eTests(args.artifacts,
                                    args.kubeconfig,
                                    args.sonobuoy_image,
                                    args.sonobuoy_version)
    sonobuoy_e2e.cleanup_cluster(sonobuoy_args)


def define_parser(parser):
    subparsers = parser.add_subparsers()

    shared_parser = argparse.ArgumentParser(add_help=False)
    shared_parser.add_argument('--artifacts',
                               default='results',
                               help='Directory where junit XML files are stored')
    shared_parser.add_argument('--kubeconfig',
                               default=os.environ.get('KUBECONFIG'),
                               help='Path to kubeconfig file')
    shared_parser.add_argument('--sonobuoy-image',
                               default='sonobuoy/sonobuoy',
                               help='Set the sonobuoy image to be used')
    shared_parser.add_argument('--sonobuoy-version',
                               default='latest',
                               help='Set the sonobuoy version to be used')

    run_parser = subparsers.add_parser('run', help='Run the tests', parents=[shared_parser])
    run_parser.add_argument('--max_failures',
                            type=int,
                            default=5,
                            help='How many failures to allow in a row while checking the test status')
    run_parser.set_defaults(func=run_tests)

    collect_parser = subparsers.add_parser('collect', help='Collect the results', parents=[shared_parser])
    collect_parser.add_argument('--collection-retries',
                                type=int,
                                default=10,
                                help='How many times to try collecting the results')
    collect_parser.set_defaults(func=collect_results)

    cleanup_parser = subparsers.add_parser('cleanup', help='Cleanup the cluster', parents=[shared_parser])
    cleanup_parser.set_defaults(func=cleanup)


if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s %(name)s: %(levelname)s: %(message)s', level='INFO')
    parser = argparse.ArgumentParser(description='Run the K8s E2E tests using sonobuoy.',
                                     epilog='Any additional args will be passed to sonobuoy.')
    define_parser(parser)
    parser.set_defaults(func=run_tests)

    args, sonobuoy_args = parser.parse_known_args()
    args.func(args, sonobuoy_args)
