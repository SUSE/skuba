import contextlib
import logging
from timeit import default_timer as timer


logger = logging.getLogger('testrunner')


@contextlib.contextmanager
def timed(description: str) -> None:
    start = timer()
    yield
    end = timer()
    logger.info("Executing {} took: {} seconds".format(description, end - start))
