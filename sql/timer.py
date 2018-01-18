# encoding=utf8
import time


class Timer(object):
    def __init__(self):
        pass

    def __enter__(self):
        self.start = time.time()
        return self

    def __exit__(self, *args):
        end = time.time()
        self.interval = end - self.start
