#!/usr/bin/env python
# encoding: utf-8

import random
import sys
from gevent import monkey, socket, joinall, spawn, ssl, sleep
monkey.patch_all()

TOTAL_NUM = 1 
ROOM_NUM = 1 
class Client(object):

    def __init__(self, addr, port, device_id, session_id):

        self.soc = socket.socket()
        self.device_id = device_id

        self.soc.connect((addr, port))
        self.soc.send("IDENTITY %s %s\r\n" % (device_id, session_id))
        self.soc.send("JOIN WORLD HelloWorld reason\r\n")
        self.soc.send("NORMAL WORLD HelloWorldreason\r\n")
        joinall([spawn(self.recv)])

    def send(self):
        f = open("/mnt/share/a.txt")
        l = f.readlines()
        sleep(5)
        while True:
            text = random.choice(l)
            self.soc.send("NORMAL %d  %s" % (random.randint(1,TOTAL_NUM)%ROOM_NUM, text))
            sleep(5)
        

    def recv(self):
        while True:
            data = self.soc.recv(1024)
            if len(data) == 0:
                sys.exit(0)
            print '**************RECV',data


if __name__ == '__main__':

    #jobs = [spawn(Client, "localhost", 8170, "GG", "41861aae627a") for x in range(1)]
    jobs = [spawn(Client, "localhost", 8170, "player1", "1") for x in range(TOTAL_NUM)]
    joinall(jobs)
