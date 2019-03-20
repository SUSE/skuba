#!/usr/bin/env python3

import argparse
import logging
import os
import subprocess
import sys
import time

logger = logging.getLogger('Sonobuoy-E2E-Tests')
DEFAULT_SLEEP = 5


def _abort(msg):
    logger.error(msg)
    sys.exit(1)


def _sonobuoy(docker_args, sonobuoy_args):
    cmd = 'docker run --rm --network=host ' \
          '-v {kubeconfig}:/root/.kube/config ' \
          '-v {artifacts}:/results ' \
          '-i {sonobuoy_image}:{sonobuoy_version} ' \
          './sonobuoy {sonobuoy_args}'.format(sonobuoy_args=sonobuoy_args, **docker_args)
    logger.info(cmd)

    proc = subprocess.run(cmd,
                          encoding='utf8',
                          shell=True,
                          stdout=subprocess.PIPE,
                          stderr=subprocess.STDOUT)

    if proc.stdout:
        logger.info(proc.stdout)

    return proc.stdout


def _start_the_tests(docker_args, sonobuoy_run_args):
    logger.info('Starting the tests')
    _sonobuoy(docker_args, 'run ' + ' '.join(sonobuoy_run_args))

    logger.info('Waiting for the tests to start...')
    start_time = time.time()
    run_time = int(time.time() - start_time)

    # Shouldn't take more than 5 min for the tests to start
    while run_time < 300:
        if 'Sonobuoy is still running' in _sonobuoy(docker_args, 'status'):
            logger.info('Tests have started!')
            return

        time.sleep(DEFAULT_SLEEP)
        run_time = int(time.time() - start_time)

    _abort('Timed out while waiting for the tests to start.')


def _wait_for_the_tests(docker_args, timeout):
    logger.info('Waiting for the tests to finish can take up to 2-3 hours...')
    start_time = time.time()
    run_time = int(time.time() - start_time)

    while run_time < timeout:
        # Check the status every two minutes
        if run_time % 120 and 'Sonobuoy is still running' not in _sonobuoy(docker_args, 'status'):
            break
        time.sleep(DEFAULT_SLEEP)
        run_time = int(time.time() - start_time)


def _collect_results(docker_args, retries, path):
    error = None
    attempts = 0

    if 'Sonobuoy has completed' in _sonobuoy(docker_args, 'status'):
        while retries > attempts:
            try:
                attempts += 1
                logger.warning('Attempting to retrieve the results {}/{}'.format(attempts, retries))
                _sonobuoy(docker_args, 'retrieve {}'.format(path))
            except:
                if retries == attempts:
                    error = 'Could not retrieve sonobuoy results'
                    break
                time.sleep(DEFAULT_SLEEP)
            else:
                break

    elif _sonobuoy(docker_args, 'status') == 'Sonobuoy is still running':
        error = 'Sonobuoy e2e tests ran out of time'
    else:
        error = 'Sonobuoy e2e tests failed'

    if error is not None:
        _abort(error)


def run(args):
    docker_args = {'artifacts': args.artifacts,
                   'kubeconfig': args.kubeconfig,
                   'sonobuoy_image': args.sonobuoy_image,
                   'sonobuoy_version': args.sonobuoy_version}
    sonobuoy_run_args = []

    if args.e2e_focus:
        sonobuoy_run_args.append('--e2e-focus {}'.format(args.e2e_focus))
    if args.e2e_skip:
        sonobuoy_run_args.append('--e2e-skip {}'.format(args.e2e_skip))

    if not os.path.isdir(args.artifacts):
        os.mkdir(args.artifacts)

    _start_the_tests(docker_args, sonobuoy_run_args)
    _wait_for_the_tests(docker_args, args.timeout * 60)
    _collect_results(docker_args, args.collection_retries, args.artifacts)
    logger.info('Done.')


def define_parser():
    parser = argparse.ArgumentParser(description='Run the K8s E2E tests using sonobuoy',
                                     formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    parser.add_argument('--artifacts',
                        default='results',
                        help='directory where junit XML files are stored')
    parser.add_argument('--collection-retries',
                        type=int,
                        default=10,
                        help='How many times to try collecting the results')
    parser.add_argument('--e2e-focus',
                        metavar='REGEX',
                        help='set the e2e tests to focus on')
    parser.add_argument('--e2e-skip',
                        metavar='REGEX',
                        help='set the e2e tests to skip')
    parser.add_argument('--kubeconfig',
                        default=os.environ.get('KUBECONFIG'),
                        help='Path to kubeconfig file')
    parser.add_argument('--sonobuoy-image',
                        default='gcr.io/heptio-images/sonobuoy',
                        help='set the sonobuoy image to be used')
    parser.add_argument('--sonobuoy-version',
                        default='latest',
                        help='set the sonobuoy version to be used')
    parser.add_argument('--timeout',
                        type=int,
                        default=180,
                        help='How long to wait for the tests to finish in minutes')

    return parser


if __name__ == '__main__':
    logging.basicConfig(format='%(asctime)s %(levelname)s: %(name)s: %(message)s', level='INFO')
    parser = define_parser()
    args = parser.parse_args()

    if not args.kubeconfig or not os.path.isfile(args.kubeconfig):
        _abort('No kubeconfig file found at {} Use the --kubeconfig option or '
               'set the environment variable KUBECONFIG'.format(args.kubeconfig))

    run(args)
