class CommandResult:
    """
    Embeds the result of running a command for easier testing of different
    output channels (stdout, stderr) and the return code of the execution.

    Delegates iterators, direct comparisons and unknown attribute retrieval
    (including function calls) to the `stdout` attribute for better testing
    ergonomics.
    """
    def __init__(self, retcode, stdout, stderr):
        self.retcode = retcode
        self.stdout = stdout
        self.stderr = stderr

    def __eq__(self, other):
        if isinstance(other, str):
            return self.stdout == other
        elif isinstance(other, self.__class__):
            return self.retcode == other.retcode and self.stdout == other.stdout and self.stderr == other.stderr
        return False

    def __contains__(self, other):
        if isinstance(other, str):
            return self.stdout.__contains__(other)
        elif isinstance(other, self.__class__):
            return self.retcode == other.retcode and self.stdout.__contains__(other.stdout) and self.stderr.__contains__(other.stderr)
        return False

    def __iter__(self):
        return self.stdout.__iter__()

    def __repr__(self):
        return f'{self.__class__.__name__}({self.retcode!r}, {self.stdout!r}, {self.stderr!r})'

    def __str__(self):
        return self.stdout

    def __int__(self):
        return int(self.__str__())

    def __getattr__(self, name):
        def forward_method(*args, **kwargs):
            return getattr(self.stdout, name)(*args, **kwargs)
        return forward_method
