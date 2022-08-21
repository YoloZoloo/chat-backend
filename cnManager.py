import mysql.connector
import os

def init_con():
    env_file = os.getcwd() + "/.env"
    with open(env_file) as f:
        USER, PASSWORD = f.readline().rstrip("\n").split(":")
        HOST, PORT = f.readline().rstrip("\n").split(":")
        f.close()
        cnx = mysql.connector.connect(user=USER, password=PASSWORD, host=HOST, port=PORT)
        return cnx