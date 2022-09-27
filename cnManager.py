import mysql.connector
import os

def init_con():
    env_file = os.getcwd() + "/.env"
    with open(env_file) as f:
        USER = f.readline().rstrip("\n")
        PASSWORD = f.readline().rstrip("\n")
        HOST = f.readline().rstrip("\n")
        PORT = f.readline().rstrip("\n")
        f.close()
        cnx = mysql.connector.connect(user=USER, password=PASSWORD, host=HOST, port=PORT)
        return cnx