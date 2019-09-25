import signal
import time

def wait(func, *args, **kwargs):

    class TimeoutError(Exception):
        pass

    timeout = kwargs.pop("wait_timeout", 5)
    delay   = kwargs.pop("wait_delay", 0)
    backoff = kwargs.pop("wait_backoff", 1)
    retries = kwargs.pop("wait_retries", 3)
    allow   = kwargs.pop("wait_allow", ())
    def _handle_timeout(signum, frame):
        raise TimeoutError()

    time.sleep(delay)
    reason=""
    for i in range(0, retries):
        signal.signal(signal.SIGALRM, _handle_timeout)
        signal.alarm(timeout)
        try:
            return func(*args, **kwargs)
        except TimeoutError:
            reason = "timeout {}s exceded".format(timeout)
        except allow as ex:
            reason = "{}: '{}'".format(ex.__class__.__name__, ex)
        finally:
            signal.alarm(0)

        time.sleep(backoff)

    raise Exception("Failed waiting for function {} due to {} after {} retries".format(func.__name__, reason, retries))

