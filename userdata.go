package main

import (
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

func cleanUserData(uid int64) bool {
	log.WithField("uid", uid).Debugln("Purging userdata...")
	_, exist := users.data[uid]
	if exist {
		os.RemoveAll(users.data[uid].userDir)
		users.mu.Lock()
		delete(users.data, uid)
		users.mu.Unlock()
		log.WithField("uid", uid).Debugln("Userdata purged from map and disk.")
		return true
	} else {
		log.WithField("uid", uid).Debugln("Userdata does not exist, do nothing.")
		return false
	}
}

func initUserData(c tele.Context, command string, state string) {
	uid := c.Sender().ID
	users.mu.Lock()
	users.data[uid] = &UserData{
		state:       state,
		userDir:     filepath.Join(dataDir, strconv.FormatInt(uid, 10)),
		command:     command,
		lineData:    &LineData{},
		stickerData: &StickerData{},
	}
	users.mu.Unlock()
	// Sanitize user work directory.
	os.RemoveAll(users.data[uid].userDir)
	os.MkdirAll(users.data[uid].userDir, 0755)
}

func getState(c tele.Context) (string, string) {
	ud, exist := users.data[c.Sender().ID]
	if exist {
		return ud.command, ud.state
	} else {
		return "", ""
	}
}

func checkState(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		state, _ := getState(c)
		// if user sent something and we are waiting for callback, user entered "nostate"
		// We should allow user entering other command during "nostate"
		if state == "" || state == "nostate" {
			log.Debugf("User %d entering command with message: %s", c.Sender().ID, c.Message().Text)
			return next(c)
		} else {
			log.Debugf("User %d already in phase: %v", c.Sender().ID, state)
			return sendInStateWarning(c)
		}
	}
}

func setState(c tele.Context, state string) {
	uid := c.Sender().ID
	users.data[uid].state = state
}

func setCommand(c tele.Context, state string) {
	uid := c.Sender().ID
	users.data[uid].command = state
}