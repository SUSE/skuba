#!/usr/bin/env python3

import argparse
import logging
import os
import subprocess
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
        error = None
        attempts = 0

        if 'Sonobuoy has completed' in self._get_status():
            while retries > attempts:
                try:
                    attempts += 1
                    logger.info(f'Attempting to retrieve the results {attempts}/{retries}')
                    retrieve = self._sonobuoy('retrieve results' + ' '.join(sonobuoy_args))
                    logger.info(retrieve)
                except SonobuoyE2eTestsError:
                    if retries == attempts:
                        error = 'Could not retrieve sonobuoy results'
                        break
                    time.sleep(self.default_sleep)
                else:
                    break
        else:
            error = 'Sonobuoy e2e tests failed'

        if error is not None:
            raise SonobuoyE2eTestsError(error)

    def run_tests(self, timeout, sonobuoy_args):
        logger.info('Getting the sonobuoy image...')
        self._pull_image()

        logger.info('Starting the tests')
        self._start_the_tests(sonobuoy_args)

        logger.info('Waiting for the tests to finish can take up to 2-3 hours...')
        self._wait_for_the_tests(timeout)

    def _get_status(self):
        status = self._sonobuoy('status')
        logger.info(status)
        return status

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
        self._sonobuoy('run ' + ' '.join(sonobuoy_args))

        logger.info('Waiting for the tests to start...')
        start_time = time.time()
        run_time = int(time.time() - start_time)

        # Shouldn't take more than 5 min for the tests to start
        while run_time < 300:
            try:
                if 'Sonobuoy is still running' in self._get_status():
                    logger.info('Tests have started!')
                    break
            except SonobuoyE2eTestsError:
                pass

            time.sleep(self.default_sleep)
            run_time = int(time.time() - start_time)

        if run_time >= 300:
            raise SonobuoyE2eTestsError('Timed out while waiting for the tests to start.')

    def _wait_for_the_tests(self, timeout):
        start_time = time.time()
        run_time = int(time.time() - start_time)
        timeout = timeout * 60

        while run_time < timeout:
            # Check the status every two minutes
            if run_time % 120:
                if 'Sonobuoy is still running' not in self._get_status():
                    break

            time.sleep(self.default_sleep)
            run_time = int(time.time() - start_time)

        if 'Sonobuoy is still running' in self._get_status():
            raise SonobuoyE2eTestsError('Sonobuoy e2e tests ran out of time')


class SonobuoyE2eTestsError(Exception):
    pass


def run_tests(args, sonobuoy_args):
    sonobuoy_e2e = SonobuoyE2eTests(args.artifacts,
                                    args.kubeconfig,
                                    args.sonobuoy_image,
                                    args.sonobuoy_version)

    sonobuoy_e2e.run_tests(args.timeout, sonobuoy_args)


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
                               default='gcr.io/heptio-images/sonobuoy',
                               help='Set the sonobuoy image to be used')
    shared_parser.add_argument('--sonobuoy-version',
                               default='latest',
                               help='Set the sonobuoy version to be used')

    run_parser = subparsers.add_parser('run', help='Run the tests', parents=[shared_parser])
    run_parser.add_argument('--timeout',
                            type=int,
                            default=180,
                            help='How long to wait for the tests to finish in minutes')
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
