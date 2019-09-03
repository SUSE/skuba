class Format:
    DOT = '\033[34m●\033[0m'
    DOT_EXIT = '\033[32m●\033[0m'
    RED = '\033[31m'
    RED_EXIT = '\033[0m'

    @staticmethod
    def alert(msg):
        return ("{}{}{}".format(Format.RED, msg, Format.RED_EXIT))
