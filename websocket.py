#!/usr/bin/env python3
# WS server example that synchronizes state across clients

import asyncio
import json
import logging
import string
import traceback
import websockets
import datetime
import cnManager
import ssl
import pathlib

logging.basicConfig()

MAPPINGDICT = dict()
RECEIVERS = set()
cnt = 0
NOT_SELECTED = -1

async def broadcastToPeers(msg, sender_id, datename, room:int , peer_id: int):
    RECEIVERS.clear()
    jsn = json.dumps({"message":msg, "senderID":sender_id, "dateName":datename, "room":room, "peerID":peer_id})
    cnx = cnManager.init_con()
    cursor = cnx.cursor()
    sSQL = "SELECT guest_id FROM chat.grouproom_m WHERE grouproom_id = %s"
    para = (int(room),)
    cursor.execute(sSQL, para)
    rslt = cursor.fetchall()
    for row in rslt:
        RECEIVERS.add(row[0])
    # Make sure data is committed to the database
    cnx.close()
    #taking user IDs out of RECEIVERS
        #From mapping, find corresponding websocket
    for recep in RECEIVERS:
        recep = int(recep)
        ws = MAPPINGDICT.get(recep)
        if ws == None:
            pass
        else:
            await ws.send(jsn)

async def notifyPeer(msg: str, sender_id: int, datename: str, peer_id: int):
    jsn = json.dumps({"message":msg, "senderID":sender_id, "dateName":datename, "room": NOT_SELECTED , "peerID":peer_id})
    pairWS = [MAPPINGDICT.get(peer_id), MAPPINGDICT.get(sender_id)]
    for ws in pairWS:
        if ws != None:  # asyncio.wait doesn't accept an empty list
            try:
                await ws.send(jsn)
            except:
                traceback.print_exc()

async def register(websocket, user_id: int):
    isExists = MAPPINGDICT.get(user_id)
    if isExists == None:
        MAPPINGDICT[user_id] = websocket
        print("registered index: " + str(MAPPINGDICT.get(user_id)))
    else:
        MAPPINGDICT[user_id] = websocket
        print("re-registered index: " + str(MAPPINGDICT.get(user_id)))


async def unregister(websocket):
    pass


async def counter(websocket, path):
    try:
        async for message in websocket:
            data = json.loads(message)
            if data["connect"] == True:
                await register(websocket, int(data["user_id"]))
            else:
                msg = data["message"]
                sender_id = int(data["senderID"])
                peer_id = int(data["peerID"])
                room = int(data["chatroom"])
                name = data["name"]
                datename = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S") + ": "+ name
                ws = MAPPINGDICT.get(sender_id)
                if ws == None:
                    await websocket.close(code=1000, reason='Websocket not found')
                    return
                if peer_id == NOT_SELECTED:
                    await broadcastToPeers(msg, sender_id, datename, room, peer_id)
                    await insertToGroup(msg, sender_id, room)
                else:
                    await notifyPeer(msg, sender_id, datename, peer_id)
                    await insertToPrivate(msg, sender_id, peer_id)
    except:
        traceback.print_exc()
    finally:
        await unregister(websocket)

async def insertToGroup(msg: str, sender_id: int, room: int):
    cnx = cnManager.init_con()
    cursor = cnx.cursor()
    iSQL = "INSERT INTO chat.groupchat_t(message, sender_id, grouproom_id) VALUES(%s, %s, %s)"
    para = (msg, sender_id, room)
    cursor.execute(iSQL, para)
    # Make sure data is committed to the database
    cnx.commit()
    cnx.close()

async def insertToPrivate(msg :str, sender_id: int, peer_id: int):
    cnx = cnManager.init_con()
    cursor = cnx.cursor()
    sSQL = "SELECT privateroom_id from chat.privateroom_m where (idA = %s AND idB = %s) OR (idB = %s AND idA = %s)"
    para = (sender_id, peer_id, sender_id, peer_id)
    cursor.execute(sSQL, para)
    rslt = cursor.fetchone()
    pRoomId = rslt[0]

    iSQL = "INSERT INTO chat.privatechat_t(message, sender_id, privateroom_id) VALUES(%s, %s, %s)"
    para = (msg, sender_id, int(pRoomId))
    cursor.execute(iSQL, para)
    # Make sure data is committed to the database
    cnx.commit()
    cnx.close()

# ssl_context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
# ssl_context.load_cert_chain('fullchain.pem', 'privkey.pem')
# print(ssl_context)
start_server = websockets.serve(counter, "0.0.0.0", 6789)
#start_server = websockets.serve(counter, "0.0.0.0", 6789, ssl=ssl_context)

asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()