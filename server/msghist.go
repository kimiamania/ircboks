//message history
package main

import (
	"code.google.com/p/go.net/websocket"
	log "github.com/ngmoco/timber"
	"labix.org/v2/mgo/bson"
)

//MessageHist represent a message history
type MessageHist struct {
	ID        bson.ObjectId `bson:"_id" json:"Id"`
	UserID    string        `bson:"userId" json:"UserId"`
	Target    string        `bson:"target"`     //message target/receiver
	Nick      string        `bson:"nick"`       //sender's nick
	Message   string        `bson:"message"`    //message content
	Timestamp int64         `bson:"timestamp"`  //server timestamp
	ReadFlag  bool          `bson:"read_flag"`  //true if this message already read by user
	ToChannel bool          `bson:"to_channel"` //true if it is message to channel
	Incoming  bool          `bson:"incoming"`   //true if it is incoming message
}

//MsgHistChannel get message history of a channel
//It will fetch message history from the latest to oldest
func MsgHistChannel(em *EndptMsg, ws *websocket.Conn) {
	channel, ok := em.GetDataString("channel")
	if !ok {
		log.Error("MsgHistChannel() null channel")
		return
	}

	log.Debug("[MsgHistChannel] userId=" + em.UserID + ".channel = " + channel)

	i := 0
	for {
		var res []MessageHist

		query := bson.M{"userId": em.UserID, "target": channel}
		err := DBQueryArr("ircboks", "msghist", query, "-timestamp", 50, 50*i, &res)
		if err != nil {
			log.Error("[MsgHistChannel]fetching channel history:" + err.Error())
			return
		}

		m := map[string]interface{}{
			"logs":    res,
			"channel": channel,
		}

		//send the result
		websocket.Message.Send(ws, jsonMarshal("msghistChannel", m))

		if len(res) == 0 || res[len(res)-1].ReadFlag == true {
			break
		}
		i = i + 1
	}
}

//MsgHistNick get message history of a nick
func MsgHistNick(em *EndptMsg, ws *websocket.Conn) {
	nick, ok := em.GetDataString("nick")
	if !ok {
		log.Error("MsgHistNick() empty nick")
		return
	}
	msgHistNick(em.UserID, nick, ws)
}

func msgHistNick(userID, nick string, ws *websocket.Conn) {
	i := 0
	query1 := bson.M{"userId": userID, "nick": nick, "to_channel": false} //message from this nick, not in channel
	query2 := bson.M{"userId": userID, "target": nick}                    //message to this nick
	query := bson.M{"$or": []bson.M{query1, query2}}

	for {
		var hists []MessageHist

		err := DBQueryArr("ircboks", "msghist", query, "-timestamp", 50, 50*i, &hists)
		if err != nil {
			log.Error("[MsgHistNick]fetching channel nick:" + err.Error())
			return
		}

		m := map[string]interface{}{
			"logs": hists,
			"nick": nick,
		}

		//send it back
		websocket.Message.Send(ws, jsonMarshal("msghistNickResp", m))

		if len(hists) == 0 || hists[len(hists)-1].ReadFlag == true {
			break
		}
		i++
	}
}

//MsgHistNicksUnread get all unread messages that is not from channel
func MsgHistNicksUnread(em *EndptMsg, ws *websocket.Conn) {
	var unreadNicks []string

	query := bson.M{"userId": em.UserID, "to_channel": false, "incoming": true, "read_flag": false}
	if err := DBSelectDistinct("ircboks", "msghist", query, "nick", &unreadNicks); err != nil {
		log.Error("MsgHistNicksUnread:selecr distinct err :" + err.Error())
		return
	}

	m := map[string]interface{}{
		"nicks": unreadNicks,
	}

	websocket.Message.Send(ws, jsonMarshal("msghistNicksUnread", m))
}

//MsgHistMarkRead mark messages readFlag as read
func MsgHistMarkRead(em *EndptMsg, ws *websocket.Conn) {
	oids := em.Args
	if len(oids) == 0 {
		log.Error("MsgHistMarkRead() empty oids")
		return
	}
	for _, oid := range oids {
		updQuery := bson.M{"$set": bson.M{"read_flag": true}}

		DBUpdateOne("ircboks", "msghist", oid, updQuery)
	}
}

//MsgHistInsert save a message to DB
func MsgHistInsert(userID, target, nick, message string, timestamp int64, readFlag, incoming bool) bson.ObjectId {
	objectID := bson.NewObjectId()
	toChannel := false
	if string(target[0]) == "#" {
		toChannel = true
	}
	doc := MessageHist{objectID, userID, target, nick, message, timestamp, readFlag, toChannel, incoming}
	err := DBInsert("ircboks", "msghist", &doc)
	if err != nil {
		log.Error("[insertMsgHistory] failed : " + err.Error())
	}
	return objectID
}
