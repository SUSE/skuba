import logging
import os


class Logger:

    def __init__(self, conf):
        pass

    @staticmethod
    def config_logger(conf, level=None):
        logging.basicConfig(
            format='%(asctime)s %(levelname)s] %(message)s',
            level=logging.DEBUG,
            datefmt='%Y-%m-%d %H:%M:%S')

        logger = logging.getLogger("testrunner")

        if conf.log.file:
            mode = 'a'
            if conf.log.overwrite:
                mode = 'w'
            log_file = os.path.join(conf.workspace, conf.log.file)
            file_handler = logging.FileHandler(log_file)
            logger.addHandler(file_handler)

        if not conf.log.quiet:
            if not level:
                level = conf.log.level
            console = logging.StreamHandler()
            console.setLevel(logging.getLevelName(level.upper()))
            logger.addHandler(console)
