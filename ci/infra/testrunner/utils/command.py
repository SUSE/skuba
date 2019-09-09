class CommandResult:
    """
    Embeds the result of running a command for easier testing of different
    output channels (stdout, stderr) and the return code of the execution.
    """
    def __init__(self, retcode, stdout, stderr):
        self.retcode = retcode
        self.stdout = stdout
        self.stderr = stderr
