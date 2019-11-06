import signal
import time
from utils.timeit import timed
import logging


logger = logging.getLogger('testrunner')


def wait(func, *args, **kwargs):

    class TimeoutError(Exception):
        pass

    timeout = kwargs.pop("wait_timeout", 0)
    delay   = kwargs.pop("wait_delay", 0)
    backoff = kwargs.pop("wait_backoff", 0)
    retries = kwargs.pop("wait_retries", 0)
    allow   = kwargs.pop("wait_allow", ())
    elapsed = kwargs.pop("wait_elapsed", 0)

    if retries > 0 and elapsed > 0:
        raise ValueError("wait_retries and wait_elapsed cannot both have a non zero value")

    if retries == 0 and elapsed == 0:
        raise ValueError("either wait_retries  or wait_elapsed must have a non zero value")

    def _handle_timeout(signum, frame):
        raise TimeoutError()

    start = int(time.time())
    attempts = 1
    reason=""

    time.sleep(delay)
    while True:
        signal.signal(signal.SIGALRM, _handle_timeout)
        signal.alarm(timeout)
        try:
            with timed(func.__name__):
                result = func(*args, **kwargs)
            return result
        except TimeoutError:
            if elapsed > 0 and int(time.time())-start >= elapsed:
               reason = "maximum wait time exceeded: {}s".format(elapsed)
               break
            reason = "timeout of {}s exceded".format(timeout)
        except allow as ex:
            reason = "{}: '{}'".format(ex.__class__.__name__, ex)
        finally:
            signal.alarm(0)

        if retries > 0 and attempts == retries:
            break

        time.sleep(backoff)

        attempts = attempts + 1

    raise Exception("Failed waiting for function {} after {} attemps due to {}".format(func.__name__, attempts, reason))

